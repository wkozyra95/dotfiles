package sway

import (
	"fmt"
)

type WorkspaceBinding struct {
	Workspace string
	Output    string
}

// WorkspacePair defines two workspaces (each pinned to its own output)
// that should be focused together. When the user switches to either
// side, the listener will switch the other side's output to the paired
// workspace.
type WorkspacePair struct {
	A WorkspaceBinding
	B WorkspaceBinding
}

// Handler is what a user supplies to Listen. Compose multiple by
// combining them with ChainHandlers.
type Handler func(l *Listener, ev Event)

// ChainHandlers returns a handler that fans an event out to every
// supplied handler in order.
func ChainHandlers(handlers ...Handler) Handler {
	return func(l *Listener, ev Event) {
		for _, h := range handlers {
			h(l, ev)
		}
	}
}

// WorkspaceSyncHandler keeps configured workspace pairs in sync
// across outputs: when the user switches to either side, the other
// side's output follows to the paired workspace.
func WorkspaceSyncHandler(pairs []WorkspacePair) Handler {
	return func(l *Listener, ev Event) {
		focus, ok := ev.(WorkspaceFocusEvent)
		if !ok || !focus.OutputChanged {
			return
		}
		if l.IsMuted() {
			log.Debugf("Muted: ignoring switch on %s/%s", focus.Output, focus.Workspace)
			return
		}
		match, target, ok := findPairTarget(pairs, focus.Workspace, focus.Output)
		if !ok {
			return
		}
		log.Infof("Pair switch: %s/%s -> %s/%s",
			match.Output, match.Workspace, target.Output, target.Workspace)
		l.ArmMute()
		l.mu.Lock()
		l.visible[target.Output] = target.Workspace
		l.mu.Unlock()
		cmd := fmt.Sprintf(
			"focus output %s; workspace --no-auto-back-and-forth %s; focus output %s",
			target.Output, target.Workspace, match.Output,
		)
		if err := l.Run(cmd); err != nil {
			log.Errorf("Failed to switch paired workspace: %v", err)
		}
	}
}

func findPairTarget(
	pairs []WorkspacePair, workspace, output string,
) (WorkspaceBinding, WorkspaceBinding, bool) {
	for _, p := range pairs {
		if p.A.Workspace == workspace && p.A.Output == output {
			return p.A, p.B, true
		}
		if p.B.Workspace == workspace && p.B.Output == output {
			return p.B, p.A, true
		}
	}
	return WorkspaceBinding{}, WorkspaceBinding{}, false
}

// MoveAppToVisibleWorkspaceHandler watches for new windows belonging
// to a specific app and moves them to whichever of the candidate
// workspaces is currently visible. Falls back to the first candidate
// if none is visible. App matching is done against both the Wayland
// app_id and the XWayland window class.
func MoveAppToVisibleWorkspaceHandler(appID string, candidates []string) Handler {
	return func(l *Listener, ev Event) {
		newEv, ok := ev.(WindowNewEvent)
		if !ok || len(candidates) == 0 {
			return
		}
		if newEv.AppID != appID && newEv.Class != appID {
			return
		}
		target := candidates[0]
		l.mu.Lock()
		for _, c := range candidates {
			for _, ws := range l.visible {
				if ws == c {
					target = c
					break
				}
			}
		}
		l.mu.Unlock()
		log.Infof("Moving %s (con_id=%d) to workspace %s", appID, newEv.ConID, target)
		cmd := fmt.Sprintf(
			"[con_id=%d] move container to workspace %s",
			newEv.ConID, target,
		)
		if err := l.Run(cmd); err != nil {
			log.Errorf("Failed to move %s: %v", appID, err)
		}
	}
}

// PortraitSplitvHandler keeps portrait-rotated outputs in a vertical
// split layout. It covers two scenarios:
//
//  1. The user focuses a workspace on a portrait output: run `splitv`
//     on the focused container so the next window stacks below it.
//  2. A new window appears on a portrait workspace (placed via a sway
//     `assign` rule, possibly while the workspace is unfocused and
//     possibly already containing other windows): run
//     `[con_id=N] layout splitv`. That changes the *parent*
//     container's layout to vertical, so any existing siblings re-flow
//     and the new window ends up below them instead of to the right.
//     (Plain `splitv` would only wrap the new window in its own
//     vertical container without rearranging existing siblings.)
func PortraitSplitvHandler() Handler {
	return func(l *Listener, ev Event) {
		switch ev := ev.(type) {
		case WorkspaceFocusEvent:
			if !ev.OutputChanged || !l.IsPortrait(ev.Output) {
				return
			}
			if err := l.Run("splitv"); err != nil {
				log.Errorf("Failed to set splitv on %s: %v", ev.Output, err)
			}
		case WindowNewEvent:
			output := OutputForCon(ev.ConID)
			if output == "" || !l.IsPortrait(output) {
				return
			}
			cmd := fmt.Sprintf("[con_id=%d] layout splitv", ev.ConID)
			if err := l.Run(cmd); err != nil {
				log.Errorf("Failed to set layout splitv around con_id=%d: %v", ev.ConID, err)
			}
		}
	}
}
