// Package server exposes api/hostd over HTTP: routing, authentication and
// serialization. What the endpoints actually do is implemented in api/hostd.
package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wkozyra95/dotfiles/api/hostd"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/notify"
)

var log = logger.NamedLogger("hostd")

const DefaultListen = ":7420"

type Config struct {
	Listen string
	// Holds the token, StateDirectory= of the systemd unit.
	StateDir string
	// Enable /files, flat directory for uploads and downloads. Disabled if empty.
	FilesDir string
	// Enable /notify: desktop notifications in the session of the user running the service.
	Notify bool
}

// Routes and handlers are in routes.go.
type server struct {
	config Config
	auth   *auth
	// nil if file transfer is disabled
	files    *hostd.FileStore
	suspend  func() error
	powerOff func() error
}

func Serve(config Config) error {
	token, err := hostd.EnsureToken(config.StateDir)
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}
	s := &server{config: config, auth: newAuth(token), suspend: hostd.Suspend, powerOff: hostd.PowerOff}
	if config.FilesDir != "" {
		if s.files, err = hostd.NewFileStore(config.FilesDir, hostd.DefaultMinFreeSpace); err != nil {
			return fmt.Errorf("files directory: %w", err)
		}
	}
	if config.Notify {
		if err := notify.CheckSession(); err != nil {
			return fmt.Errorf("notifications: %w", err)
		}
	}
	httpServer := &http.Server{
		Addr:    config.Listen,
		Handler: s.router(),
		// file transfers lift those for themselves
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Infof("Listening on %s", config.Listen)
	return httpServer.ListenAndServe()
}
