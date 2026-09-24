package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Requests and responses are authenticated with HMAC-SHA256 over a shared
// token that never travels over the network, so a request sent to the wrong
// machine (the same LAN IP on a foreign network) leaks nothing reusable.
//
// The signature is base64(HMAC(token, lines joined with "\n")), where lines are
//
//	request:   "request", METHOD, URI, time, nonce, hex(sha256(body))
//	multipart: "request", METHOD, URI, time, nonce, "UNSIGNED-PAYLOAD"
//	response:  "response", nonce, status, hex(sha256(body))
//
// The HMAC key is the token exactly as `mycli hostd state` prints it (ASCII
// bytes of the hex string), URI is the request target as sent (path + query),
// time is unix seconds, nonce is chosen by the client. The response signature
// lets the client tell this server from whatever else might answer on the same
// address.
//
// File bodies are too large to hash up front, so the body of multipart
// requests (Content-Type: multipart/*) is not covered by the signature at all.
// Nothing in such a body can be trusted more than the network it came through,
// anything that matters has to be in the URI. The client also has to make sure
// it talks to the real server (a verified response to some other request)
// before it starts sending a file. Downloads carry the body hash in
// HeaderContentHash and the response signature covers that value.
const (
	HeaderTime        = "X-Hostd-Time"
	HeaderNonce       = "X-Hostd-Nonce"
	HeaderSignature   = "X-Hostd-Signature"
	HeaderContentHash = "X-Hostd-Content-SHA256"

	unsignedPayload = "UNSIGNED-PAYLOAD"

	maxClockSkew = 30 * time.Second
	minNonceLen  = 16
	maxNonceLen  = 64
	maxBodySize  = 64 * 1024
)

type auth struct {
	token []byte
	now   func() time.Time

	// Nonces of accepted requests, kept for as long as their timestamp could
	// still pass the clock skew check.
	mutex  sync.Mutex
	nonces map[string]time.Time
}

func newAuth(token string) *auth {
	return &auth{token: []byte(token), now: time.Now, nonces: map[string]time.Time{}}
}

func (a *auth) sign(parts ...string) string {
	mac := hmac.New(sha256.New, a.token)
	mac.Write([]byte(strings.Join(parts, "\n")))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (a *auth) requestSignature(method, uri, timestamp, nonce string, body []byte) string {
	return a.sign("request", method, uri, timestamp, nonce, bodyHash(body))
}

func (a *auth) multipartSignature(method, uri, timestamp, nonce string) string {
	return a.sign("request", method, uri, timestamp, nonce, unsignedPayload)
}

func (a *auth) responseSignature(nonce string, status int, contentHash string) string {
	return a.sign("response", nonce, strconv.Itoa(status), contentHash)
}

// signResponse has to be called before the status is written. Responses to
// requests without a usable nonce stay unsigned.
func (a *auth) signResponse(w http.ResponseWriter, r *http.Request, status int, contentHash string) {
	if nonce := r.Header.Get(HeaderNonce); isValidNonce(nonce) {
		w.Header().Set(HeaderSignature, a.responseSignature(nonce, status, contentHash))
	}
}

func bodyHash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func isValidNonce(nonce string) bool {
	if len(nonce) < minNonceLen || len(nonce) > maxNonceLen {
		return false
	}
	for _, c := range nonce {
		isAlnum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !isAlnum && !strings.ContainsRune("+/=-_", c) {
			return false
		}
	}
	return true
}

// claimNonce returns false if the nonce was already used.
func (a *auth) claimNonce(nonce string, now time.Time) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for key, usedAt := range a.nonces {
		if now.Sub(usedAt) > 2*maxClockSkew {
			delete(a.nonces, key)
		}
	}
	if _, used := a.nonces[nonce]; used {
		return false
	}
	a.nonces[nonce] = now
	return true
}

// verify checks everything but the body, expectedSignature gets the already
// validated time and nonce headers.
func (a *auth) verify(r *http.Request, expectedSignature func(timestamp, nonce string) string) error {
	nonce := r.Header.Get(HeaderNonce)
	if !isValidNonce(nonce) {
		return errors.New("missing or invalid nonce")
	}
	timestamp := r.Header.Get(HeaderTime)
	unixSeconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("missing or invalid timestamp")
	}
	if !hmac.Equal([]byte(expectedSignature(timestamp, nonce)), []byte(r.Header.Get(HeaderSignature))) {
		return errors.New("invalid signature")
	}
	now := a.now()
	if skew := now.Sub(time.Unix(unixSeconds, 0)); skew > maxClockSkew || skew < -maxClockSkew {
		return errors.New("clock skew too large")
	}
	// Only nonces of correctly signed requests are stored, so they can not
	// be piled up by someone without the token.
	if !a.claimNonce(nonce, now) {
		return errors.New("nonce already used")
	}
	return nil
}

func (a *auth) reject(w http.ResponseWriter, r *http.Request, err error) {
	log.Warnf("%s %s from %s rejected: %v", r.Method, r.URL.Path, r.RemoteAddr, err)
	// Rejections are signed too, a client with a wrong clock still needs to
	// know that serverTime comes from the real server.
	respond(a, w, r, http.StatusUnauthorized, errorResponse{Error: err.Error(), ServerTime: a.now().Unix()})
}

// authenticate is the middleware in front of every route. Which of the two
// request signatures is expected depends only on the format of the body.
func (a *auth) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isMultipart(r) {
			// verified before anything is read from the body
			err := a.verify(r, func(timestamp, nonce string) string {
				return a.multipartSignature(r.Method, r.RequestURI, timestamp, nonce)
			})
			if err != nil {
				a.reject(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))
		if err != nil {
			respond(a, w, r, http.StatusRequestEntityTooLarge, errorResponse{Error: "body too large"})
			return
		}
		err = a.verify(r, func(timestamp, nonce string) string {
			return a.requestSignature(r.Method, r.RequestURI, timestamp, nonce, body)
		})
		if err != nil {
			a.reject(w, r, err)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

func isMultipart(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && strings.HasPrefix(mediaType, "multipart/")
}
