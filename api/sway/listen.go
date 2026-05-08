package sway

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/wkozyra95/dotfiles/env"
	swayexec "github.com/wkozyra95/dotfiles/utils/exec"
)

type workspaceEvent struct {
	Change  string `json:"change"`
	Current struct {
		Name   string `json:"name"`
		Output string `json:"output"`
	} `json:"current"`
}

// muteWindow drops any workspace focus events received within a short
// window after we issue a programmatic switch. Sway emits several
// intermediate focus events for a chained `focus output ..., workspace ...,
// focus output ...` command (including one for the source workspace when
// focus returns), and reacting to them would cause an infinite loop.
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

// ListenWorkspacePairs subscribes to sway's workspace events and keeps the
// configured workspace pairs in sync across outputs. The function blocks
// until the subscribe process exits.
func ListenWorkspacePairs(pairs []env.SwayWorkspacePair) error {
	if len(pairs) == 0 {
		log.Info("No sway workspace pairs configured, nothing to do")
		return nil
	}

	mute := newMuteWindow(200 * time.Millisecond)
	// visible[output] = workspace currently shown on that output. Used to
	// distinguish a real workspace change from a plain focus shift across
	// outputs (sway fires change=focus for both).
	visible := map[string]string{}

	cmd := exec.Command("swaymsg", "-t", "subscribe", "-m", `["workspace"]`)
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
		var ev workspaceEvent
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			log.Debugf("Failed to decode sway event: %v", err)
			continue
		}
		if ev.Change != "focus" {
			continue
		}
		output := ev.Current.Output
		workspace := ev.Current.Name
		prev, seen := visible[output]
		visible[output] = workspace
		if seen && prev == workspace {
			// Output's visible workspace didn't change — this is just a
			// focus shift between outputs, not a workspace switch.
			continue
		}
		if mute.muted() {
			log.Debugf("Muted: ignoring switch on %s/%s", output, workspace)
			continue
		}
		match, target, ok := findPairTarget(pairs, workspace, output)
		if !ok {
			continue
		}
		log.Infof("Pair switch: %s/%s -> %s/%s",
			match.Output, match.Workspace, target.Output, target.Workspace)
		mute.arm()
		visible[target.Output] = target.Workspace
		if err := switchPairedWorkspace(match, target); err != nil {
			log.Errorf("Failed to switch paired workspace: %v", err)
		}
	}
	return scanner.Err()
}

// findPairTarget returns the binding the user just focused (match) and the
// binding on the other monitor that should follow (target).
func findPairTarget(
	pairs []env.SwayWorkspacePair, workspace, output string,
) (env.SwayWorkspaceBinding, env.SwayWorkspaceBinding, bool) {
	for _, p := range pairs {
		if p.A.Workspace == workspace && p.A.Output == output {
			return p.A, p.B, true
		}
		if p.B.Workspace == workspace && p.B.Output == output {
			return p.B, p.A, true
		}
	}
	return env.SwayWorkspaceBinding{}, env.SwayWorkspaceBinding{}, false
}

// switchPairedWorkspace focuses the target output, switches its workspace,
// and returns focus to the originating output so the user does not lose
// their place.
func switchPairedWorkspace(source, target env.SwayWorkspaceBinding) error {
	command := fmt.Sprintf(
		"focus output %s; workspace --no-auto-back-and-forth %s; focus output %s",
		target.Output, target.Workspace, source.Output,
	)
	return swayexec.Command().Args("swaymsg", command).Run()
}
