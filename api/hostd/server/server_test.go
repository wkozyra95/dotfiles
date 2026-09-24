package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wkozyra95/dotfiles/api/hostd"
)

const testToken = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

var testNow = time.Unix(1789820000, 0)

type testServer struct {
	*server
	// power actions that reached the host
	actions []string
}

type testRequest struct {
	method string
	uri    string
	body   string
	nonce  string
	time   time.Time
	token  string
	// multipart requests are signed without the body
	contentType string
}

func newTestServer(t *testing.T, withFiles bool) *testServer {
	s := &testServer{server: &server{auth: newAuth(testToken), push: hostd.NewPushTokenStore(t.TempDir())}}
	s.auth.now = func() time.Time { return testNow }
	s.suspend = func() error { s.actions = append(s.actions, "suspend"); return nil }
	s.powerOff = func() error { s.actions = append(s.actions, "poweroff"); return nil }
	if withFiles {
		files, err := hostd.NewFileStore(t.TempDir(), 0)
		assert.Nil(t, err)
		s.files = files
	}
	return s
}

var nonceCounter = 0

func (s *testServer) send(req testRequest) *httptest.ResponseRecorder {
	if req.nonce == "" {
		nonceCounter++
		req.nonce = fmt.Sprintf("test-nonce-%010d", nonceCounter)
	}
	if req.time.IsZero() {
		req.time = testNow
	}
	if req.token == "" {
		req.token = testToken
	}
	client := newAuth(req.token)
	timestamp := strconv.FormatInt(req.time.Unix(), 10)
	signature := client.requestSignature(req.method, req.uri, timestamp, req.nonce, []byte(req.body))
	if strings.HasPrefix(req.contentType, "multipart/") {
		signature = client.multipartSignature(req.method, req.uri, timestamp, req.nonce)
	}
	r := httptest.NewRequest(req.method, req.uri, strings.NewReader(req.body))
	if req.contentType != "" {
		r.Header.Set("Content-Type", req.contentType)
	}
	r.Header.Set(HeaderTime, timestamp)
	r.Header.Set(HeaderNonce, req.nonce)
	r.Header.Set(HeaderSignature, signature)
	w := httptest.NewRecorder()
	s.router().ServeHTTP(w, r)
	return w
}

func multipartBody(fileName string, content string) (string, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("ignored", "form field")
	part, _ := writer.CreateFormFile("file", fileName)
	_, _ = part.Write([]byte(content))
	_ = writer.Close()
	return body.String(), writer.FormDataContentType()
}

func (s *testServer) upload(uri string, content string) *httptest.ResponseRecorder {
	// the name in the body is not signed, it must not be used for anything
	body, contentType := multipartBody("../name-from-the-body.txt", content)
	return s.send(testRequest{method: "PUT", uri: uri, body: body, contentType: contentType})
}

func TestStatusWithSignedResponse(t *testing.T) {
	s := newTestServer(t, false)
	w := s.send(testRequest{method: "GET", uri: "/status", nonce: "status-nonce-0001"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(
		t,
		s.auth.responseSignature("status-nonce-0001", http.StatusOK, bodyHash(w.Body.Bytes())),
		w.Header().Get(HeaderSignature),
	)
	status := statusResponse{}
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &status))
	hostname, _ := os.Hostname()
	assert.Equal(t, hostname, status.Hostname)
	assert.Equal(t, []string{"push"}, status.Features)

	w = newTestServer(t, true).send(testRequest{method: "GET", uri: "/status"})
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &status))
	assert.Equal(t, []string{"push", "files"}, status.Features)
}

func TestPushToken(t *testing.T) {
	s := newTestServer(t, false)
	registered := func() []string {
		devices, err := s.push.List()
		assert.Nil(t, err)
		names := []string{}
		for _, device := range devices {
			names = append(names, device.Name+"="+device.Token)
		}
		return names
	}

	w := s.send(testRequest{method: "POST", uri: "/push-token", body: `{"token":"fcm:token-1","device":"Phone"}`})
	assert.Equal(t, http.StatusOK, w.Code)
	response := pushDeviceResponse{}
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "Phone", response.Device)
	assert.NotZero(t, response.RegisteredAt)
	w = s.send(testRequest{method: "POST", uri: "/push-token", body: `{"token":"fcm:token-2","device":"Tablet"}`})
	assert.Equal(t, http.StatusOK, w.Code)
	// rotated token
	w = s.send(testRequest{method: "POST", uri: "/push-token", body: `{"token":"fcm:token-3","device":"Phone"}`})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"Tablet=fcm:token-2", "Phone=fcm:token-3"}, registered())

	rejected := []string{
		`not json`,
		`{"device":"Phone"}`,
		`{"token":"","device":"Phone"}`,
		`{"token":"fcm:token-4"}`,
		`{"token":"fcm:token-4","device":"  "}`,
		`{"token":"has space","device":"Phone"}`,
		`{"token":"fcm:token-4","device":"Phone","extra":1}`,
	}
	for _, body := range rejected {
		w = s.send(testRequest{method: "POST", uri: "/push-token", body: body})
		assert.Equal(t, http.StatusBadRequest, w.Code, body)
	}
	assert.Equal(t, []string{"Tablet=fcm:token-2", "Phone=fcm:token-3"}, registered())

	w = s.send(testRequest{method: "DELETE", uri: "/push-token", body: `{"token":"fcm:token-2"}`})
	assert.Equal(t, http.StatusOK, w.Code)
	w = s.send(testRequest{method: "DELETE", uri: "/push-token", body: `{"token":"fcm:token-2"}`})
	assert.Equal(t, http.StatusOK, w.Code)
	w = s.send(testRequest{method: "DELETE", uri: "/push-token", body: `{"device":"Phone"}`})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, []string{"Phone=fcm:token-3"}, registered())

	// unsigned requests never reach the store
	w = s.send(
		testRequest{
			method: "POST",
			uri:    "/push-token",
			body:   `{"token":"fcm:token-5","device":"Evil"}`,
			token:  "wrong",
		},
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, []string{"Phone=fcm:token-3"}, registered())
}

func TestRejectedRequests(t *testing.T) {
	s := newTestServer(t, false)

	w := s.send(testRequest{method: "POST", uri: "/suspend", token: "wrong"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = s.send(
		testRequest{method: "POST", uri: "/suspend", time: testNow.Add(-time.Minute), nonce: "skewed-nonce-0001"},
	)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(
		t,
		s.auth.responseSignature("skewed-nonce-0001", http.StatusUnauthorized, bodyHash(w.Body.Bytes())),
		w.Header().Get(HeaderSignature),
	)
	rejection := errorResponse{}
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &rejection))
	assert.Equal(t, testNow.Unix(), rejection.ServerTime)

	w = s.send(testRequest{method: "POST", uri: "/suspend", nonce: "replayed-nonce-01"})
	assert.Equal(t, http.StatusAccepted, w.Code)
	w = s.send(testRequest{method: "POST", uri: "/suspend", nonce: "replayed-nonce-01"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// unsigned requests learn nothing, not even which routes exist
	for _, uri := range []string{"/suspend", "/unknown"} {
		unsigned := httptest.NewRecorder()
		s.router().ServeHTTP(unsigned, httptest.NewRequest("POST", uri, nil))
		assert.Equal(t, http.StatusUnauthorized, unsigned.Code)
		assert.Equal(t, "", unsigned.Header().Get(HeaderSignature))
	}

	assert.Equal(t, []string{"suspend"}, s.actions)
}

func TestSignatureCoversRequest(t *testing.T) {
	s := newTestServer(t, true)
	timestamp := strconv.FormatInt(testNow.Unix(), 10)
	sendSigned := func(method, uri, body, contentType, nonce, signature string) int {
		r := httptest.NewRequest(method, uri, strings.NewReader(body))
		r.Header.Set("Content-Type", contentType)
		r.Header.Set(HeaderTime, timestamp)
		r.Header.Set(HeaderNonce, nonce)
		r.Header.Set(HeaderSignature, signature)
		w := httptest.NewRecorder()
		s.router().ServeHTTP(w, r)
		return w.Code
	}

	// signature of a status request reused for poweroff
	nonce := "tampered-nonce-01"
	signature := s.auth.requestSignature("GET", "/status", timestamp, nonce, nil)
	assert.Equal(t, http.StatusUnauthorized, sendSigned("POST", "/poweroff", "", "", nonce, signature))

	// which signature is expected depends on the content type, so changing
	// the content type of a request in transit does not get its body ignored
	body, contentType := multipartBody("a.txt", "x")
	nonce = "tampered-nonce-02"
	signature = s.auth.requestSignature("PUT", "/files/a.txt", timestamp, nonce, []byte(body))
	assert.Equal(t, http.StatusUnauthorized, sendSigned("PUT", "/files/a.txt", body, contentType, nonce, signature))
	nonce = "tampered-nonce-03"
	signature = s.auth.multipartSignature("PUT", "/files/a.txt", timestamp, nonce)
	assert.Equal(t, http.StatusUnauthorized, sendSigned("PUT", "/files/a.txt", body, "", nonce, signature))
	// and the multipart signature still covers the URI
	nonce = "tampered-nonce-04"
	assert.Equal(t, http.StatusUnauthorized, sendSigned("PUT", "/files/b.txt", body, contentType, nonce, signature))

	assert.Empty(t, s.actions)
	files, _ := s.files.List()
	assert.Empty(t, files)
}

func TestPowerActions(t *testing.T) {
	s := newTestServer(t, false)
	assert.Equal(t, http.StatusAccepted, s.send(testRequest{method: "POST", uri: "/suspend"}).Code)
	assert.Equal(t, http.StatusAccepted, s.send(testRequest{method: "POST", uri: "/poweroff"}).Code)
	assert.Equal(t, http.StatusMethodNotAllowed, s.send(testRequest{method: "GET", uri: "/poweroff"}).Code)
	assert.Equal(t, http.StatusNotFound, s.send(testRequest{method: "GET", uri: "/unknown"}).Code)
	assert.Equal(t, []string{"suspend", "poweroff"}, s.actions)

	s.suspend = func() error { return errors.New("logind CanSuspend returned s \"challenge\"") }
	assert.Equal(t, http.StatusInternalServerError, s.send(testRequest{method: "POST", uri: "/suspend"}).Code)
}

func TestFiles(t *testing.T) {
	s := newTestServer(t, true)

	w := s.upload("/files/with%20space.txt", "hello")
	assert.Equal(t, http.StatusCreated, w.Code)
	uploaded := fileResponse{}
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &uploaded))
	assert.Equal(t, "with space.txt", uploaded.Name)
	assert.Equal(t, int64(5), uploaded.Size)

	w = s.send(testRequest{method: "GET", uri: "/files"})
	assert.Equal(t, http.StatusOK, w.Code)
	list := fileListResponse{}
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &list))
	assert.Equal(t, []fileResponse{uploaded}, list.Files)
	assert.Greater(t, list.FreeBytes, uint64(0))

	w = s.send(testRequest{method: "GET", uri: "/files/with%20space.txt", nonce: "download-nonce-01"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello", w.Body.String())
	assert.Equal(t, "5", w.Header().Get("Content-Length"))
	assert.Equal(t, bodyHash([]byte("hello")), w.Header().Get(HeaderContentHash))
	assert.Equal(
		t,
		s.auth.responseSignature("download-nonce-01", http.StatusOK, bodyHash([]byte("hello"))),
		w.Header().Get(HeaderSignature),
	)

	// the body has to be multipart with a file in it
	assert.Equal(t, http.StatusBadRequest, s.send(testRequest{method: "PUT", uri: "/files/raw.txt", body: "raw"}).Code)
	w = s.send(testRequest{
		method: "PUT", uri: "/files/raw.txt", body: "--x--\r\n", contentType: "multipart/form-data; boundary=x",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// cut short, the closing boundary never comes
	body, contentType := multipartBody("a.txt", "content")
	w = s.send(testRequest{method: "PUT", uri: "/files/cut.txt", body: body[:len(body)-20], contentType: contentType})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// errors of the file store
	assert.Equal(t, http.StatusConflict, s.upload("/files/with%20space.txt", "other").Code)
	assert.Equal(t, http.StatusCreated, s.upload("/files/with%20space.txt?overwrite=1", "other").Code)
	for _, uri := range []string{"/files/..%2Foutside.txt", "/files/.hidden", "/files/a%2Fb"} {
		assert.Equal(t, http.StatusBadRequest, s.upload(uri, "x").Code, uri)
		assert.Equal(t, http.StatusBadRequest, s.send(testRequest{method: "GET", uri: uri}).Code, uri)
	}
	assert.Equal(t, http.StatusNotFound, s.send(testRequest{method: "GET", uri: "/files/missing.txt"}).Code)

	assert.Equal(t, http.StatusOK, s.send(testRequest{method: "DELETE", uri: "/files/with%20space.txt"}).Code)
	assert.Equal(t, http.StatusNotFound, s.send(testRequest{method: "DELETE", uri: "/files/with%20space.txt"}).Code)
}

func TestFilesDisabled(t *testing.T) {
	s := newTestServer(t, false)
	assert.Equal(t, http.StatusNotFound, s.send(testRequest{method: "GET", uri: "/files"}).Code)
	assert.Equal(t, http.StatusNotFound, s.send(testRequest{method: "GET", uri: "/files/a.txt"}).Code)
	assert.Equal(t, http.StatusNotFound, s.upload("/files/a.txt", "x").Code)
}

// stubNotifySend puts a fake notify-send first on PATH and returns the file it
// appends its arguments to, one call per line.
func stubNotifySend(t *testing.T) string {
	dir := t.TempDir()
	calls := filepath.Join(dir, "calls")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + calls + "\n"
	assert.Nil(t, os.WriteFile(filepath.Join(dir, "notify-send"), []byte(script), 0o755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return calls
}

func notifySendCalls(calls string) []string {
	content, err := os.ReadFile(calls)
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
}

func TestNotify(t *testing.T) {
	s := newTestServer(t, false)
	s.config.Notify = true
	calls := stubNotifySend(t)

	w := s.send(testRequest{method: "GET", uri: "/status"})
	status := statusResponse{}
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &status))
	assert.Equal(t, []string{"push", "notify"}, status.Features)

	w = s.send(testRequest{method: "POST", uri: "/notify", body: `{"title":"Hi","message":"plain text"}`})
	assert.Equal(t, http.StatusAccepted, w.Code)
	w = s.send(testRequest{
		method: "POST", uri: "/notify",
		body: `{"title":"Link","message":"look","action":{"type":"share-url","url":"https://example.com/a?b=1"}}`,
	})
	assert.Equal(t, http.StatusAccepted, w.Code)
	w = s.send(testRequest{
		method: "POST", uri: "/notify",
		body: `{"title":"Text","message":"","action":{"type":"share-text","text":"secret words"}}`,
	})
	assert.Equal(t, http.StatusAccepted, w.Code)
	// notify-send with an action runs in the background, so the order of the calls varies
	expected := []string{
		"--app-name=mycli -- Hi plain text",
		"--app-name=mycli --wait --action=default=Open -- Link look",
		"--app-name=mycli --wait --action=default=Copy -- Text ",
	}
	assert.Eventually(t, func() bool { return len(notifySendCalls(calls)) == 3 }, 5*time.Second, 10*time.Millisecond)
	assert.ElementsMatch(t, expected, notifySendCalls(calls))

	rejected := []string{
		`not json`,
		`{"message":"no title"}`,
		`{"title":"  ","message":"blank title"}`,
		`{"title":"x","unknown":1}`,
		`{"title":"x","action":{"type":"share-url","url":"ftp://example.com"}}`,
		`{"title":"x","action":{"type":"share-url","url":"https://example.com","text":"both"}}`,
		`{"title":"x","action":{"type":"share-text","text":""}}`,
		`{"title":"x","action":{"type":"share-file","text":"x"}}`,
		`{"title":"` + strings.Repeat("a", 201) + `"}`,
	}
	for _, body := range rejected {
		w = s.send(testRequest{method: "POST", uri: "/notify", body: body})
		assert.Equal(t, http.StatusBadRequest, w.Code, body)
	}
	assert.ElementsMatch(t, expected, notifySendCalls(calls))
}

func TestNotifyDisabled(t *testing.T) {
	w := newTestServer(t, false).send(testRequest{method: "POST", uri: "/notify", body: `{"title":"Hi"}`})
	assert.Equal(t, http.StatusNotFound, w.Code)
}
