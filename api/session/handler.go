package session

import (
	"bytes"
	"strconv"

	"github.com/wkozyra95/dotfiles/api/sway"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

// waybarSessionSignal is the SIGRTMIN offset the waybar custom/session module
// listens on (must match its "signal" in configs/waybar/config).
const waybarSessionSignal = "-RTMIN+8"

// ListenerHandler keeps the session state and bar indicator in sync with the
// live windows. On window close it clears the active tag of any group that became
// empty; it refreshes the waybar session indicator on workspace focus changes
// and window closes (closing the last window of a workspace emits no workspace
// event, so the bar must be nudged explicitly).
func ListenerHandler() sway.Handler {
	return func(_ *sway.Listener, ev sway.Event) {
		switch ev.(type) {
		case sway.WindowCloseEvent:
			reconcileActiveTags()
			refreshBar()
		case sway.WorkspaceFocusEvent:
			refreshBar()
		}
	}
}

// reconcileActiveTags drops the active tag of every group that no longer has any
// windows.
func reconcileActiveTags() {
	if err := getStateManager().RunGuarded(func(s *SessionState) error {
		for key := range s.Active {
			primary, convErr := strconv.Atoi(key)
			if convErr != nil {
				continue
			}
			if groupOf(primary).windowCount() == 0 {
				delete(s.Active, key)
			}
		}
		return nil
	}); err != nil {
		log.Errorf("Failed to reconcile active tags: %v", err)
	}
}

// refreshBar signals waybar to re-evaluate the session indicator.
func refreshBar() {
	var stdout, stderr bytes.Buffer
	// Ignore the error: pkill exits non-zero when waybar is not running.
	_ = exec.Command().WithBufout(&stdout, &stderr).
		Args("pkill", waybarSessionSignal, "waybar").Run()
}
