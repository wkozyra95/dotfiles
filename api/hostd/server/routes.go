package server

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/wkozyra95/dotfiles/api/hostd"
	"github.com/wkozyra95/dotfiles/utils/notify"
)

func (s *server) router() http.Handler {
	router := chi.NewRouter()
	json := func(handler jsonHandler) http.HandlerFunc { return jsonRoute(s.auth, handler) }

	// Requests are authenticated before anything about the API is revealed,
	// that includes unknown routes.
	router.Use(s.auth.authenticate)
	router.NotFound(json(notFound))
	router.MethodNotAllowed(json(methodNotAllowed))

	router.Get("/status", json(s.handleStatus))
	router.Post("/suspend", json(s.handleSuspend))
	router.Post("/poweroff", json(s.handlePowerOff))
	if s.config.Notify {
		router.Post("/notify", json(s.handleNotify))
	}
	if s.files != nil {
		router.Get("/files", json(s.handleFileList))
		router.With(noDeadlines).Put("/files/{name}", json(s.handleFileUpload))
		router.With(noDeadlines).Get("/files/{name}", s.handleFileDownload)
		router.Delete("/files/{name}", json(s.handleFileDelete))
	}
	return router
}

func (s *server) handleStatus(*http.Request) (int, any) {
	status, err := hostd.ReadStatus()
	if err != nil {
		return errorResult(err)
	}
	response := statusResponse{
		Hostname:      status.Hostname,
		UptimeSeconds: int64(status.Uptime.Seconds()),
		Features:      []string{},
	}
	if s.files != nil {
		response.Features = append(response.Features, "files")
	}
	if s.config.Notify {
		response.Features = append(response.Features, "notify")
	}
	return http.StatusOK, response
}

func (s *server) handleSuspend(*http.Request) (int, any) {
	if err := s.suspend(); err != nil {
		return errorResult(err)
	}
	return http.StatusAccepted, struct{}{}
}

func (s *server) handlePowerOff(*http.Request) (int, any) {
	if err := s.powerOff(); err != nil {
		return errorResult(err)
	}
	return http.StatusAccepted, struct{}{}
}

// handleNotify answers as soon as the notification is accepted, it stays on the
// screen (and the action may run) long after that.
func (s *server) handleNotify(r *http.Request) (int, any) {
	request := notifyRequest{}
	if err := decodeJSON(r, &request); err != nil {
		return http.StatusBadRequest, errorResponse{Error: err.Error()}
	}
	notification := notify.Notification{Title: request.Title, Message: request.Message}
	if request.Action != nil {
		notification.Action = request.Action.Action
	}
	if err := notification.Validate(); err != nil {
		return errorResult(err)
	}
	notify.Notify(notification)
	return http.StatusAccepted, struct{}{}
}

// fileNameFromRequestPath returns the name from /files/{name}. It is taken from the decoded
// path, so an escaped slash reaches the validation in hostd.FileStore as a
// regular one.
func fileNameFromRequestPath(r *http.Request) string {
	return strings.TrimPrefix(r.URL.Path, "/files/")
}

func (s *server) handleFileList(*http.Request) (int, any) {
	files, err := s.files.List()
	if err != nil {
		return errorResult(err)
	}
	response := fileListResponse{Files: []fileResponse{}}
	for _, file := range files {
		response.Files = append(response.Files, newFileResponse(file))
	}
	if response.FreeBytes, err = s.files.FreeSpace(); err != nil {
		return errorResult(err)
	}
	return http.StatusOK, response
}

// handleFileUpload takes a multipart body and stores its first file part. The
// body is not covered by the request signature, so the name comes from the
// URI and the file name of the part is ignored.
func (s *server) handleFileUpload(r *http.Request) (int, any) {
	content, err := firstFileFromMultipart(r)
	if err != nil {
		return http.StatusBadRequest, errorResponse{Error: err.Error()}
	}
	overwrite := r.URL.Query().Get("overwrite") == "1"
	// Content-Length includes the multipart framing, close enough for a size hint
	info, err := s.files.Save(fileNameFromRequestPath(r), content, r.ContentLength, overwrite)
	if err != nil {
		return errorResult(err)
	}
	return http.StatusCreated, newFileResponse(info)
}

func (s *server) handleFileDelete(r *http.Request) (int, any) {
	if err := s.files.Delete(fileNameFromRequestPath(r)); err != nil {
		return errorResult(err)
	}
	return http.StatusOK, struct{}{}
}

// handleFileDownload streams the file, so unlike JSON responses the signature
// covers the hash declared in HeaderContentHash instead of a buffered body.
func (s *server) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	file, err := s.files.Open(fileNameFromRequestPath(r))
	if err != nil {
		jsonRoute(s.auth, func(*http.Request) (int, any) { return errorResult(err) })(w, r)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(file.Info.Size, 10))
	w.Header().Set(HeaderContentHash, file.Hash)
	s.auth.signResponse(w, r, http.StatusOK, file.Hash)
	w.WriteHeader(http.StatusOK)
	written, err := io.CopyN(w, file, file.Info.Size)
	if err != nil {
		log.Warnf("Download of %s interrupted after %d of %d bytes [%v]", file.Info.Name, written, file.Info.Size, err)
		return
	}
	log.Infof("%s %s from %s -> %d", r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK)
}
