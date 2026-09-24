package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wkozyra95/dotfiles/api/hostd"
	"github.com/wkozyra95/dotfiles/utils/notify"
)

// jsonHandler returns the status code and the value to encode as the body.
type jsonHandler func(r *http.Request) (int, any)

func jsonRoute(a *auth, handler jsonHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, response := handler(r)
		log.Infof("%s %s from %s -> %d", r.Method, r.URL.Path, r.RemoteAddr, status)
		respond(a, w, r, status, response)
	}
}

// respond writes a signed JSON response.
func respond(a *auth, w http.ResponseWriter, r *http.Request, status int, response any) {
	body, err := json.Marshal(response)
	if err != nil {
		status = http.StatusInternalServerError
		body = []byte(`{"error":"failed to encode response"}`)
	}
	a.signResponse(w, r, status, bodyHash(body))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// errorResult maps errors of the hostd package to a status code and a body.
func errorResult(err error) (int, any) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, hostd.ErrInvalidFileName), errors.Is(err, hostd.ErrIncompleteUpload),
		errors.Is(err, hostd.ErrInvalidPushToken), errors.Is(err, notify.ErrInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, hostd.ErrFileNotFound):
		status = http.StatusNotFound
	case errors.Is(err, hostd.ErrFileExists):
		status = http.StatusConflict
	case errors.Is(err, hostd.ErrNoSpace):
		status = http.StatusInsufficientStorage
	default:
		log.Error(err)
	}
	return status, errorResponse{Error: err.Error()}
}

const maxJSONBody = 64 << 10

// decodeJSON reads a small JSON request body, unknown fields are rejected.
func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxJSONBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	return nil
}

func notFound(*http.Request) (int, any) {
	return http.StatusNotFound, errorResponse{Error: "not found"}
}

func methodNotAllowed(*http.Request) (int, any) {
	return http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"}
}

// firstFileFromMultipart streams the body, the part is valid until the body is closed.
func firstFileFromMultipart(r *http.Request) (io.Reader, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, errors.New("multipart body is required")
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			return nil, errors.New("missing file part")
		}
		if part.FileName() != "" {
			return part, nil
		}
	}
}

// noDeadlines lifts the server wide timeouts, transfers take as long as they take.
func noDeadlines(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controller := http.NewResponseController(w)
		_ = controller.SetReadDeadline(time.Time{})
		_ = controller.SetWriteDeadline(time.Time{})
		next.ServeHTTP(w, r)
	})
}
