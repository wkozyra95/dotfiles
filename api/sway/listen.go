package sway

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"sync"
	"time"

	swayexec "github.com/wkozyra95/dotfiles/utils/exec"
)

// Event is the parsed sum type passed to the user-supplied handler.
type Event interface{ isSwayEvent() }

type WorkspaceFocusEvent struct {
	Output    string
	Workspace string
	// Prev is the previously visible workspace on Output, or "" if unknown.
	Prev string
	// OutputChanged is true when Output's visible workspace actually
	// changed; false for plain focus shifts between outputs (sway emits
	// change=focus for those too).
	OutputChanged bool
}

type WindowNewEvent struct {
	ConID int64
	AppID string
	Class string
	Title string
}

type WindowFocusEvent struct {
	ConID int64
	AppID string
	Class string
}

func (WorkspaceFocusEvent) isSwayEvent() {}
func (WindowNewEvent) isSwayEvent()      {}
func (WindowFocusEvent) isSwayEvent()    {}

type outputInfo struct {
	transform string
}

// Listener owns shared state (visible workspace per output, output
// transforms, the mute window) and exposes the helpers handlers need.
type Listener struct {
	mu      sync.Mutex
	visible map[string]string
	outputs map[string]outputInfo
	mute    *muteWindow
}

func (l *Listener) VisibleWorkspace(output string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.visible[output]
}

func (l *Listener) IsPortrait(output string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	t := l.outputs[output].transform
	return t == "90" || t == "270"
}

// ArmMute starts the suppression window. Subsequent workspace-focus
// events are reported with no special flag, but handlers can call
// IsMuted to decide whether to skip reacting (used to avoid feedback
// loops from programmatic workspace switches).
func (l *Listener) ArmMute() { l.mute.arm() }

func (l *Listener) IsMuted() bool { return l.mute.muted() }

// Run executes a swaymsg command string (e.g. "move container to workspace 6").
func (l *Listener) Run(command string) error {
	return swayexec.Command().Args("swaymsg", command).Run()
}

// muteWindow drops events received within a short window after a
// programmatic switch. Sway emits several intermediate focus events
// for chained `focus output ..., workspace ..., focus output ...`
// commands and reacting to them would cause an infinite loop.
type muteWindow struct {
	mu       sync.Mutex
	deadline time.Time
	duration time.Duration
}

func newMuteWindow(d time.Duration) *muteWindow {
	return &muteWindow{duration: d}
}

func (m *muteWindow) arm() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deadline = time.Now().Add(m.duration)
}

func (m *muteWindow) muted() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return time.Now().Before(m.deadline)
}

type rawWorkspaceEvent struct {
	Change  string `json:"change"`
	Current struct {
		Name   string `json:"name"`
		Output string `json:"output"`
	} `json:"current"`
}

type rawWindowEvent struct {
	Change    string `json:"change"`
	Container struct {
		ID               int64  `json:"id"`
		AppID            string `json:"app_id"`
		Name             string `json:"name"`
		WindowProperties struct {
			Class string `json:"class"`
		} `json:"window_properties"`
	} `json:"container"`
}

type rawOutput struct {
	Name      string `json:"name"`
	Transform string `json:"transform"`
}

func loadOutputs() (map[string]outputInfo, error) {
	var stdout, stderr bytes.Buffer
	err := swayexec.Command().
		WithBufout(&stdout, &stderr).
		Args("swaymsg", "-t", "get_outputs", "-r").Run()
	if err != nil {
		return nil, err
	}
	var raw []rawOutput
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return nil, err
	}
	out := make(map[string]outputInfo, len(raw))
	for _, r := range raw {
		out[r.Name] = outputInfo{transform: r.Transform}
	}
	return out, nil
}

// Listen subscribes to sway workspace and window events, parses them
// into Event values, and invokes handle for each. Blocks until the
// subscribe process exits.
func Listen(handle func(*Listener, Event)) error {
	outputs, err := loadOutputs()
	if err != nil {
		log.Errorf("Failed to load sway outputs: %v", err)
		outputs = map[string]outputInfo{}
	}
	l := &Listener{
		visible: map[string]string{},
		outputs: outputs,
		mute:    newMuteWindow(200 * time.Millisecond),
	}

	cmd := exec.Command("swaymsg", "-t", "subscribe", "-m", `["workspace","window"]`)
	stdout, pipeErr := cmd.StdoutPipe()
	if pipeErr != nil {
		return pipeErr
	}
	if startErr := cmd.Start(); startErr != nil {
		return startErr
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if ev, ok := parseWorkspace(line); ok {
			l.mu.Lock()
			prev, seen := l.visible[ev.Output]
			l.visible[ev.Output] = ev.Workspace
			l.mu.Unlock()
			ev.Prev = prev
			ev.OutputChanged = !seen || prev != ev.Workspace
			handle(l, ev)
			continue
		}
		if ev, ok := parseWindow(line); ok {
			handle(l, ev)
			continue
		}
	}
	return scanner.Err()
}

func parseWorkspace(line []byte) (WorkspaceFocusEvent, bool) {
	var raw rawWorkspaceEvent
	if err := json.Unmarshal(line, &raw); err != nil || raw.Change == "" {
		return WorkspaceFocusEvent{}, false
	}
	if raw.Current.Name == "" || raw.Current.Output == "" {
		return WorkspaceFocusEvent{}, false
	}
	if raw.Change != "focus" {
		return WorkspaceFocusEvent{}, false
	}
	return WorkspaceFocusEvent{
		Output:    raw.Current.Output,
		Workspace: raw.Current.Name,
	}, true
}

func parseWindow(line []byte) (Event, bool) {
	var raw rawWindowEvent
	if err := json.Unmarshal(line, &raw); err != nil || raw.Change == "" {
		return nil, false
	}
	if raw.Container.ID == 0 {
		return nil, false
	}
	switch raw.Change {
	case "new":
		return WindowNewEvent{
			ConID: raw.Container.ID,
			AppID: raw.Container.AppID,
			Class: raw.Container.WindowProperties.Class,
			Title: raw.Container.Name,
		}, true
	case "focus":
		return WindowFocusEvent{
			ConID: raw.Container.ID,
			AppID: raw.Container.AppID,
			Class: raw.Container.WindowProperties.Class,
		}, true
	}
	return nil, false
}
