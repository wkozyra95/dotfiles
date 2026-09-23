// Package notify shows desktop notifications through notify-send. Used by mycli
// commands to report problems and by hostd for messages sent from the phone.
package notify

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("notify")

type Urgency string

const (
	Low Urgency = "low"
	// Zero value, the daemon's default
	Normal   Urgency = "normal"
	Critical Urgency = "critical"
)

// Categories with their own look in the dunst config (nix/hm-modules/dunst.nix)
const (
	CategorySuccess = "success"
	CategoryWarning = "warning"
	CategoryError   = "error"
)

type Notification struct {
	Title   string
	Message string
	// Optional, matched by the rules of the notification daemon (see the dunst
	// config in nix/hm-modules/dunst.nix)
	Urgency  Urgency
	Category string
	// Overrides the timeout of the urgency and category rules of the daemon,
	// zero keeps them
	Timeout time.Duration
	// Offered as the default action of the notification, nil for none. Waiting
	// for the click happens in the background and needs a long-lived process.
	Action Action
}

// Action is one of OpenURL and CopyText; the unexported methods keep the set
// closed, each variant knows how to present, check and run itself.
type Action interface {
	// Label of the button shown by the notification daemon
	label() string
	validate() error
	run() error
}

// ActionOpenURL opens the link with the default application.
type ActionOpenURL struct {
	URL string
}

// ActionCopyText copies the text to the clipboard.
type ActionCopyText struct {
	Text string
}

var ErrInvalid = errors.New("invalid notification")

const (
	maxTitleLen   = 200
	maxMessageLen = 4000
	maxURLLen     = 2048
	maxTextLen    = 16 << 10
)

// Validate checks a notification built from untrusted input.
func (n Notification) Validate() error {
	if strings.TrimSpace(n.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalid)
	}
	if err := checkText("title", n.Title, maxTitleLen); err != nil {
		return err
	}
	if err := checkText("message", n.Message, maxMessageLen); err != nil {
		return err
	}
	switch n.Urgency {
	case "", Low, Normal, Critical:
	default:
		return fmt.Errorf("%w: unsupported urgency %q", ErrInvalid, n.Urgency)
	}
	if err := checkText("category", n.Category, maxTitleLen); err != nil {
		return err
	}
	if n.Action != nil {
		return n.Action.validate()
	}
	return nil
}

func (ActionOpenURL) label() string { return "Open" }

func (a ActionOpenURL) validate() error {
	if err := checkText("url", a.URL, maxURLLen); err != nil {
		return err
	}
	parsed, err := url.Parse(a.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%w: url must be http(s)", ErrInvalid)
	}
	return nil
}

// run opens the link with the default application. hostd is sandboxed, so the
// browser is started as a transient unit of the user's systemd manager, outside
// of the sandbox. (The desktop portal refuses callers whose /proc entries it can
// not inspect, which is the case for the service.)
func (a ActionOpenURL) run() error {
	log.Infof("Opening %s", a.URL)
	if err := a.open(); err != nil {
		Notify(Notification{Title: "Failed to open link", Message: err.Error(), Urgency: Critical})
		return err
	}
	Notify(Notification{Title: "Launching", Message: a.URL, Category: CategorySuccess, Timeout: 2 * time.Second})
	return nil
}

func (a ActionOpenURL) open() error {
	// resolved here, the user manager has a different PATH
	xdgOpen, err := exec.LookPath("xdg-open")
	if err != nil {
		return err
	}
	// xdg-open needs its helpers (xdg-mime) to look up the default browser, the
	// user manager's PATH does not have them
	cmd := command(
		"systemd-run", "--user", "--collect", "--quiet", "--setenv=PATH="+os.Getenv("PATH"),
		"--", xdgOpen, a.URL,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("systemd-run xdg-open: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (ActionCopyText) label() string { return "Copy" }

func (a ActionCopyText) validate() error {
	if strings.TrimSpace(a.Text) == "" {
		return fmt.Errorf("%w: text is required", ErrInvalid)
	}
	return checkText("text", a.Text, maxTextLen)
}

// run hands the text to wl-copy, which forks a process that keeps serving the
// clipboard. Its output must not be captured, or Wait would hang on the pipe
// until that process ends.
func (a ActionCopyText) run() error {
	log.Info("Copying text to the clipboard")
	cmd := command("wl-copy")
	cmd.Stdin = strings.NewReader(a.Text)
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("wl-copy: %w", err)
		Notify(Notification{Title: "Failed to copy", Message: err.Error(), Urgency: Critical})
		return err
	}
	Notify(Notification{Title: "Copied", Message: a.Text, Category: CategorySuccess, Timeout: 2 * time.Second})
	return nil
}

func checkText(field string, value string, maxLen int) error {
	if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		return fmt.Errorf("%w: %s is not valid text", ErrInvalid, field)
	}
	if len(value) > maxLen {
		return fmt.Errorf("%w: %s is longer than %d bytes", ErrInvalid, field, maxLen)
	}
	return nil
}

// Notify shows the notification, failures are logged. With an Action it returns
// as soon as the notification is on the screen and runs the action in the
// background when the user picks it.
func Notify(n Notification) {
	args := []string{"--app-name=mycli"}
	if n.Urgency != "" {
		args = append(args, "--urgency="+string(n.Urgency))
	}
	if n.Category != "" {
		args = append(args, "--category="+n.Category)
	}
	if n.Timeout > 0 {
		args = append(args, "--expire-time="+strconv.FormatInt(n.Timeout.Milliseconds(), 10))
	}
	if n.Action == nil {
		args = append(args, "--", n.Title, n.Message)
		if out, err := command("notify-send", args...).CombinedOutput(); err != nil {
			log.Errorf("notify-send: %v: %s", err, strings.TrimSpace(string(out)))
		}
		return
	}
	// notify-send blocks until the notification is closed and prints the
	// action the user picked
	args = append(args, "--wait", "--action=default="+n.Action.label(), "--", n.Title, n.Message)
	cmd := command("notify-send", args...)
	picked := &bytes.Buffer{}
	cmd.Stdout = picked
	if err := cmd.Start(); err != nil {
		log.Errorf("notify-send: %v", err)
		return
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Errorf("notify-send: %v", err)
			return
		}
		if strings.TrimSpace(picked.String()) != "default" {
			return
		}
		if err := n.Action.run(); err != nil {
			log.Errorf("Action of %q: %v", n.Title, err)
		}
	}()
}

// CheckSession fails when the session bus of this user is not reachable, e.g. a
// systemd unit that hides /run/user. notify-send would otherwise hang forever.
func CheckSession() error {
	address := sessionEnv()["DBUS_SESSION_BUS_ADDRESS"]
	socket, isSocket := strings.CutPrefix(address, "unix:path=")
	if !isSocket {
		return nil
	}
	if _, err := os.Stat(socket); err != nil {
		return fmt.Errorf("session bus socket: %w", err)
	}
	return nil
}

// command runs a tool of the desktop session. hostd is started by systemd
// without the session environment, so the sockets are pointed at explicitly
// when the variables are missing; sway names its first socket wayland-1.
func command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	for name, value := range sessionEnv() {
		cmd.Env = append(cmd.Env, name+"="+value)
	}
	return cmd
}

func sessionEnv() map[string]string {
	runtimeDir := envOrDefault("XDG_RUNTIME_DIR", fmt.Sprintf("/run/user/%d", os.Getuid()))
	return map[string]string{
		"XDG_RUNTIME_DIR":          runtimeDir,
		"DBUS_SESSION_BUS_ADDRESS": envOrDefault("DBUS_SESSION_BUS_ADDRESS", "unix:path="+runtimeDir+"/bus"),
		"WAYLAND_DISPLAY":          envOrDefault("WAYLAND_DISPLAY", "wayland-1"),
	}
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
