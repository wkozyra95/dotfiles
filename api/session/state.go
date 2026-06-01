package session

import (
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/persistentstate"
)

var log = logger.NamedLogger("session")

// statePath holds the session manager state (the stashed/hidden projects).
// Runtime-only, like the launcher state.
const statePath = "/tmp/mycli-session.json"

// StashedWorkspace records a single workspace that was renamed away and parked
// off-screen. Only its slot (0 = primary/main, 1 = partner/secondary) is kept:
// restore is relative to the focused group, so the original workspace number and
// output don't matter.
type StashedWorkspace struct {
	Slot int `json:"slot"`
}

// StashedSession is a hidden project occupying a unit (a workspace pair such as
// 2+6 / 3+7, or a single standalone workspace).
type StashedSession struct {
	Name       string             `json:"name"`
	Workspaces []StashedWorkspace `json:"workspaces"`
}

// SessionState is the persisted session manager state.
type SessionState struct {
	// Sessions are the stashed (hidden) projects, keyed by name.
	Sessions map[string]StashedSession `json:"sessions"`
	// Active maps a unit key (its primary workspace number, as a string) to the
	// name of the project currently live in that unit. Used to display the
	// session name on the bar and to stash without re-prompting.
	Active map[string]string `json:"active"`
}

func ensureDefault(s *SessionState) *SessionState {
	if s == nil {
		s = &SessionState{}
	}
	if s.Sessions == nil {
		s.Sessions = map[string]StashedSession{}
	}
	if s.Active == nil {
		s.Active = map[string]string{}
	}
	return s
}

func getStateManager() persistentstate.StateManager[SessionState] {
	return persistentstate.GetStateManager(statePath, "mycli-session", ensureDefault)
}
