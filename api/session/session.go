// Package session implements a sway "session manager": a project's windows can
// be stashed (hidden but left running) and later restored, so you can switch
// between projects without closing anything.
//
// # Groups
//
// Work is organized into groups (see group.go). A group is either a pair of
// workspaces — 2+6 or 3+7 (primary + partner) — or a single standalone
// workspace (1, 4, 5, ...). One project occupies one group; in a pair the
// editor sits on the primary workspace and terminals on the partner, usually on
// a second monitor.
//
// # Workspace naming
//
// Live (visible) workspaces keep their plain sway names/numbers: "2", "6", ...
// The name of the project currently live in a group is NOT encoded in the
// workspace name — it is tracked in state (Active, keyed by the group's primary
// workspace number) and shown on the bar. That keeps `workspace number N`
// navigation working and the bar uncluttered.
//
// While a project is stashed, each of its workspaces is renamed to
// "<project>:<slot>" (0 = primary/main, 1 = partner/secondary; e.g.
// "smelter:0"). The number it was stashed from is irrelevant — restore is
// slot-relative to the focused group — so only the slot is encoded. Because the
// name no longer starts with a digit sway assigns it num -1: it leaves the
// numeric namespace, so `workspace number N` can't reach it and it won't collide
// with a fresh numbered workspace.
//
// Two reserved name patterns appear transiently:
//   - "__evict:<num>" — an empty, auto-created workspace squatting on a number
//     a restore needs is renamed out of the way; being empty, sway destroys it
//     once it loses focus (see evictNumber).
//   - "__park:<name>" — see "Park output" below.
//
// # Park output (keeping stashed windows alive and off-bar)
//
// The bars are per-output, so stashed workspaces are parked on a headless
// output created at runtime via `create_output` (named "HEADLESS-N"): no
// visible bar lists it, yet its windows keep running. Workspaces are moved
// there with a workspace-criteria move ([workspace="..."] move workspace to
// output ...) so focus is not disturbed. When sway auto-creates a replacement
// numbered workspace on the park output it is swept to "__park:<old>" so it
// can't hijack `workspace number N` (see sway.SweepParkOutput).
//
// # Stash (doStashGroup)
//
// For each populated workspace N of the group: rename "N" -> "<project>:<slot>"
// (preserving its full split/tab layout) and move it to the park output. Only
// the slot is recorded; that is all a restore needs. Empty workspaces are
// skipped.
//
// # Restore (doRestore)
//
// Restore is relative to the FOCUSED group, not the project's original
// workspaces — restoring while focused on 2+6 lands the project on 2 and 6 even
// if it was stashed from 3+7. Each stashed workspace maps onto the target group
// by its stored slot (primary->primary, partner->partner), so a partner-only
// stash still lands on the partner. For each target the parked "<project>:<slot>"
// is renamed back to the plain target number (reclaiming the number, layout
// intact) after evicting any
// empty squatter; if several stashed workspaces map to one target (a pair
// restored into a standalone group) the extras are merged in with
// `move container to workspace number`. Primary goes to the focused output,
// partner to the second monitor (see workspaceGroup.outputs).
//
// # State and the bar
//
// State lives at /tmp/mycli-session.json (runtime only; Reset clears it at sway
// startup since parked workspaces don't survive a restart):
//   - Sessions: the stashed projects, keyed by name, each with its workspaces.
//   - Active:   primary-workspace-number -> live project name, for the bar.
//
// The bar shows the focused group's Active name and is signal-driven: the sway
// listener (ListenerHandler) nudges waybar (SIGRTMIN+8) on workspace focus and
// window close, and clears an Active tag once its group becomes empty.
package session

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/sway"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/menu"
)

// stashCurrentLabel is the synthetic entry in the switch picker that stashes the
// current group without restoring another session.
const stashCurrentLabel = "▸ stash current"

// closeCurrentLabel is the synthetic entry in the switch picker that closes the
// current group's windows (discarding them) without restoring another session.
const closeCurrentLabel = "▸ close current"

// ErrCanceled is returned when the user dismisses an interactive prompt. The
// command layer treats it as a silent no-op (no notification).
var ErrCanceled = errors.New("canceled")

// stashName is the workspace name a workspace is renamed to while its project is
// hidden, e.g. ("smelter", 0) -> "smelter:0" (slot 0 = main, 1 = secondary). The
// name is a prefix, so it drops out of sway's numeric namespace.
func stashName(session string, slot int) string {
	return fmt.Sprintf("%s:%d", session, slot)
}

// doStashGroup hides a group's windows: each populated workspace is renamed to
// "<name>:<slot>" and parked on the headless output. Returns what was stashed.
func doStashGroup(name string, u workspaceGroup) []StashedWorkspace {
	park, parkErr := sway.EnsureParkOutput()
	if parkErr != nil {
		log.Errorf("Failed to ensure park output: %v", parkErr)
		return nil
	}
	curByNum := map[int]sway.Workspace{}
	if current, err := sway.GetWorkspaces(); err == nil {
		for _, w := range current {
			curByNum[w.Num] = w
		}
	}
	stashed := []StashedWorkspace{}
	for _, n := range u.workspaces() {
		if sway.WorkspaceWindowCountByNum(n) == 0 {
			continue
		}
		cur, ok := curByNum[n]
		if !ok {
			continue
		}
		slot := slotOf(n)
		target := stashName(name, slot)
		// cur.Name is the live name, possibly labelled ("2:smelter").
		if err := sway.RenameWorkspace(cur.Name, target); err != nil {
			log.Errorf("Failed to rename %s: %v", cur.Name, err)
			continue
		}
		stashed = append(stashed, StashedWorkspace{Slot: slot})
		if err := sway.MoveWorkspaceToOutput(target, park); err != nil {
			log.Errorf("Failed to park %s: %v", target, err)
		}
	}
	sway.SweepParkOutput()
	return stashed
}

// closeGroupWindows kills every window on the group's workspaces, discarding the
// group rather than stashing it. Used when the user opts to close the current
// session instead of naming/stashing it.
func closeGroupWindows(u workspaceGroup) {
	for _, n := range u.workspaces() {
		sway.KillWorkspaceWindowsByNum(n)
	}
}

// evictNumber renames an existing (empty, auto-created) workspace occupying the
// given number out of the numeric namespace, so a restore can reclaim it.
func evictNumber(num int) {
	name := strconv.Itoa(num)
	if !sway.WorkspaceExists(name) {
		return
	}
	if err := sway.RenameWorkspace(name, fmt.Sprintf("__evict:%d", num)); err != nil {
		log.Errorf("Failed to evict workspace %d: %v", num, err)
	}
}

// doRestore places a stashed session into the target group (relative to the
// focused group, not the original workspaces). The stashed workspaces, ordered
// by slot, map onto the target group's workspaces: primary->primary,
// partner->partner. If the target group is standalone (or has fewer workspaces),
// the surplus stashed workspaces are merged into the last target workspace.
func doRestore(sess StashedSession, target workspaceGroup, primaryOut, partnerOut string) {
	targets := target.workspaces()

	stashed := append([]StashedWorkspace{}, sess.Workspaces...)
	sort.Slice(stashed, func(i, j int) bool { return stashed[i].Slot < stashed[j].Slot })

	claimed := map[int]bool{}
	for _, sw := range stashed {
		// Map by the workspace's slot (primary vs partner), not its position in
		// the list, so a partner workspace whose primary was empty still restores
		// to the target partner rather than the primary.
		j := min(sw.Slot, len(targets)-1)
		ti := targets[j]
		name := stashName(sess.Name, sw.Slot)
		if !claimed[ti] {
			// Claim the target number by renaming the parked workspace, which
			// preserves its layout. Place primary on the focused output and the
			// partner on the second monitor.
			out := primaryOut
			if j == 1 {
				out = partnerOut
			}
			evictNumber(ti)
			if err := sway.MoveWorkspaceToOutput(name, out); err != nil {
				log.Errorf("Failed to move %s to %s: %v", name, out, err)
			}
			if err := sway.RenameWorkspace(name, strconv.Itoa(ti)); err != nil {
				log.Errorf("Failed to rename %s to %d: %v", name, ti, err)
			}
			claimed[ti] = true
		} else {
			// Target already claimed: merge this workspace's windows into it.
			if err := sway.Command(
				fmt.Sprintf("[workspace=%q] move container to workspace number %d", name, ti),
			); err != nil {
				log.Errorf("Failed to merge %s into workspace %d: %v", name, ti, err)
			}
		}
	}
	sway.SweepParkOutput()
}

// prepareGroupLayout makes the group's workspaces exist on their outputs and
// leaves focus on the primary (real) workspace.
func prepareGroupLayout(u workspaceGroup, primaryOut, partnerOut string) {
	sway.EnsureRealFocus()
	if !u.standalone() && partnerOut != "" {
		if err := sway.FocusOutput(partnerOut); err != nil {
			log.Errorf("Failed to focus output %s: %v", partnerOut, err)
		}
		if err := sway.FocusWorkspaceNum(u.partner); err != nil {
			log.Errorf("Failed to focus workspace %d: %v", u.partner, err)
		}
	}
	if primaryOut != "" {
		if err := sway.FocusOutput(primaryOut); err != nil {
			log.Errorf("Failed to focus output %s: %v", primaryOut, err)
		}
	}
	if err := sway.FocusWorkspaceNum(u.primary); err != nil {
		log.Errorf("Failed to focus workspace %d: %v", u.primary, err)
	}
}

// pickTemplate shows the template picker and returns the chosen template.
func pickTemplate(ctx context.Context, u workspaceGroup) (env.SessionTemplate, error) {
	var templates []env.SessionTemplate
	if ctx.EnvironmentConfig.SessionTemplates != nil {
		templates = ctx.EnvironmentConfig.SessionTemplates()
	}
	if len(templates) == 0 {
		return env.SessionTemplate{}, errors.New("no session templates configured")
	}
	names := make([]string, len(templates))
	byName := map[string]env.SessionTemplate{}
	for i, tpl := range templates {
		label := tpl.Name
		if label == "" {
			label = tpl.ID
		}
		names[i] = label
		byName[label] = tpl
	}
	label := strconv.Itoa(u.primary)
	if !u.standalone() {
		label = fmt.Sprintf("%d,%d", u.primary, u.partner)
	}
	choice, picked := menu.Select(fmt.Sprintf("Launch session -> workspace %s:", label), names)
	if !picked {
		return env.SessionTemplate{}, ErrCanceled
	}
	return byName[choice], nil
}

// launchTemplate spawns each of a template's tasks as a plain terminal routed to
// the right workspace via its --class. These are ordinary windows (not
// supervised services): if the command exits, the terminal simply closes. When
// overrideCwd is non-empty it replaces each task's Cwd.
func launchTemplate(tpl env.SessionTemplate, u workspaceGroup, overrideCwd string) {
	for _, task := range tpl.Tasks {
		ws := u.primary
		if task.Slot == 1 {
			ws = u.partner
		}
		cwd := task.Cwd
		if overrideCwd != "" {
			cwd = overrideCwd
		}
		args := []string{"alacritty", "--class", fmt.Sprintf("workspace%d", ws)}
		if cwd != "" {
			args = append(args, "--working-directory", cwd)
		}
		if len(task.Args) > 0 {
			args = append(args, "-e")
			args = append(args, task.Args...)
		}
		if _, err := exec.Command().Args(args...).Start(); err != nil {
			log.Errorf("Failed to launch template task %s: %v", task.ID, err)
		}
	}
}

// nameIsFree reports whether a stash name is unused.
func nameIsFree(name string) bool {
	state, err := getStateManager().GetState()
	if err != nil {
		return true
	}
	state = ensureDefault(state)
	_, exists := state.Sessions[name]
	return !exists
}

// stashNameFor resolves the name to stash the group under: its tracked live
// name if it has one, otherwise it prompts. Returns ErrCanceled if the prompt is
// dismissed, or an error if the chosen name is already taken. Resolving the name
// has no side effects, so callers can do it up front (before other prompts) and
// only commit the stash later.
func stashNameFor(u workspaceGroup) (string, error) {
	if name := u.activeName(); name != "" {
		return name, nil
	}
	typed := menu.Prompt("Project name (stashed): ")
	if typed == "" {
		return "", ErrCanceled
	}
	if !nameIsFree(typed) {
		return "", fmt.Errorf("a stashed project named %q already exists", typed)
	}
	return typed, nil
}

// stashCurrentGroup hides the current occupant of the group under name (only
// called when the group has windows; name comes from stashNameFor).
func stashCurrentGroup(u workspaceGroup, name string) error {
	stashed := doStashGroup(name, u)
	if err := getStateManager().RunGuarded(func(s *SessionState) error {
		s.Sessions[name] = StashedSession{Name: name, Workspaces: stashed}
		delete(s.Active, u.key())
		return nil
	}); err != nil {
		return err
	}
	log.Infof("Stashed %q (%d workspaces)", name, len(stashed))
	return nil
}

// New launches a template into the focused group, prompting for the project name
// at creation. The current group (if occupied) is stashed only AFTER the template
// is picked, named and prepared, so a cancel or setup failure leaves the current
// windows untouched.
func New(ctx context.Context) error {
	sway.EnsureRealFocus()
	ws, ok := sway.FocusedWorkspace()
	if !ok {
		return errors.New("could not determine focused workspace")
	}
	u := groupOf(ws.Num)
	primaryOut, partnerOut := u.outputs(ws.Num, ws.Output)

	// If the current unit must be stashed, settle its name first — so an unnamed
	// project is named up front, before choosing/naming the new one. Resolving
	// has no side effects, so a later cancel still leaves the current windows
	// untouched.
	oldName := ""
	if u.windowCount() > 0 {
		resolved, nameErr := stashNameFor(u)
		if nameErr != nil {
			return nameErr
		}
		oldName = resolved
	}

	tpl, err := pickTemplate(ctx, u)
	if err != nil {
		return err
	}
	name := menu.Prompt("Project name: ")
	if name == "" {
		return ErrCanceled
	}
	// A template may prepare a working directory from the project name (e.g.
	// create a git worktree); its returned path overrides each task's Cwd.
	cwd := ""
	if tpl.Prepare != nil {
		prepared, prepareErr := tpl.Prepare(name)
		if prepareErr != nil {
			return fmt.Errorf("template setup failed: %w", prepareErr)
		}
		cwd = prepared
	}

	// Committed: free the group (under the name resolved above), lay it out, launch.
	if oldName != "" {
		if err := stashCurrentGroup(u, oldName); err != nil {
			return err
		}
	}
	prepareGroupLayout(u, primaryOut, partnerOut)
	launchTemplate(tpl, u, cwd)
	if err := getStateManager().RunGuarded(func(s *SessionState) error {
		s.Active[u.key()] = name
		return nil
	}); err != nil {
		log.Errorf("Failed to record active session name: %v", err)
	}
	refreshBar()
	return nil
}

// Switch restores a stashed project into the currently focused group (a pair
// places primary/partner on the two workspaces; a standalone target merges the
// stashed windows into the one workspace), stashing that group's current
// occupant first (prompting for a name only if it has windows).
func Switch() error {
	sway.EnsureRealFocus()
	ws, wsOk := sway.FocusedWorkspace()
	if !wsOk {
		return errors.New("could not determine focused workspace")
	}
	target := groupOf(ws.Num)
	primaryOut, partnerOut := target.outputs(ws.Num, ws.Output)

	// Settle the focused group's tag before switching away from it: an empty group
	// drops its (now stale) tag. An occupied group keeps its windows until the
	// picker choice commits; if it is unnamed its name is resolved at stash time
	// by stashNameFor (or it is discarded via the "close current" entry), so a
	// cancel here leaves the current windows untouched.
	hasWindows := target.windowCount() > 0
	if !hasWindows {
		if err := getStateManager().RunGuarded(func(s *SessionState) error {
			delete(s.Active, target.key())
			return nil
		}); err != nil {
			log.Errorf("Failed to clear active session name: %v", err)
		}
		refreshBar()
	}

	state, stateErr := getStateManager().GetState()
	if stateErr != nil {
		return stateErr
	}
	state = ensureDefault(state)
	candidates := []string{}
	for name := range state.Sessions {
		candidates = append(candidates, name)
	}
	// Offer to stash or close the current group (without restoring anything) when
	// it has windows, so you can free it even with no stashed sessions to switch
	// to. Listed last so the stashed sessions stay at the top.
	if hasWindows {
		candidates = append(candidates, stashCurrentLabel, closeCurrentLabel)
	}
	if len(candidates) == 0 {
		return errors.New("no stashed sessions and nothing to stash")
	}
	choice, ok := menu.Select("Switch to session:", candidates)
	if !ok {
		return ErrCanceled
	}
	switch choice {
	case stashCurrentLabel:
		name, err := stashNameFor(target)
		if err != nil {
			return err
		}
		if err := stashCurrentGroup(target, name); err != nil {
			return err
		}
		prepareGroupLayout(target, primaryOut, partnerOut)
		return nil
	case closeCurrentLabel:
		closeGroupWindows(target)
		if err := getStateManager().RunGuarded(func(s *SessionState) error {
			delete(s.Active, target.key())
			return nil
		}); err != nil {
			log.Errorf("Failed to clear active session name: %v", err)
		}
		refreshBar()
		prepareGroupLayout(target, primaryOut, partnerOut)
		return nil
	}
	sess, ok := state.Sessions[choice]
	if !ok || len(sess.Workspaces) == 0 {
		return fmt.Errorf("session %q is not stashed", choice)
	}

	// Free the focused (target) group before restoring into it.
	if hasWindows {
		name, err := stashNameFor(target)
		if err != nil {
			return err
		}
		if err := stashCurrentGroup(target, name); err != nil {
			return err
		}
	}

	doRestore(sess, target, primaryOut, partnerOut)
	if err := sway.FocusWorkspaceNum(target.primary); err != nil {
		log.Errorf("Failed to focus workspace %d: %v", target.primary, err)
	}
	// Drop the stash record and track the restored project as live in this group.
	if err := getStateManager().RunGuarded(func(s *SessionState) error {
		delete(s.Sessions, choice)
		s.Active[target.key()] = choice
		return nil
	}); err != nil {
		return err
	}
	log.Infof("Restored session %q into group %d", choice, target.primary)
	return nil
}

// List prints the tracked stashed sessions and, separately, everything that is
// actually offscreen — every workspace currently living on a non-real (headless
// park) output, with its window count. The latter reflects raw sway state, so
// it surfaces anything untracked: orphaned parks, "__park:*" placeholders or
// "__evict:*" leftovers that no longer correspond to a session record.
func List() error {
	state, stateErr := getStateManager().GetState()
	if stateErr != nil {
		return stateErr
	}
	state = ensureDefault(state)
	if len(state.Sessions) == 0 {
		log.Info("No stashed sessions")
	} else {
		log.Info("Stashed sessions:")
		for name, sess := range state.Sessions {
			log.Infof("  %s (%d workspaces)", name, len(sess.Workspaces))
		}
	}

	workspaces, err := sway.GetWorkspaces()
	if err != nil {
		return err
	}
	realOutputs := map[string]bool{}
	if outs, err := sway.RealOutputs(); err == nil {
		for _, o := range outs {
			realOutputs[o] = true
		}
	}
	offscreen := false
	for _, w := range workspaces {
		if realOutputs[w.Output] {
			continue
		}
		if !offscreen {
			log.Info("Offscreen workspaces:")
			offscreen = true
		}
		log.Infof("  [%s] on %s (%d windows)", w.Name, w.Output, sway.WorkspaceWindowCount(w.Name))
	}
	if !offscreen {
		log.Info("Nothing offscreen")
	}
	return nil
}

// Close kills the windows of a stashed session, then forgets it.
func Close(name string) error {
	manager := getStateManager()
	state, stateErr := manager.GetState()
	if stateErr != nil {
		return stateErr
	}
	state = ensureDefault(state)
	if name == "" {
		candidates := []string{}
		for n := range state.Sessions {
			candidates = append(candidates, n)
		}
		if len(candidates) == 0 {
			return errors.New("no stashed sessions to close")
		}
		chosen, ok := menu.Select("Close session:", candidates)
		if !ok {
			return ErrCanceled
		}
		name = chosen
	}
	sess, ok := state.Sessions[name]
	if !ok {
		return fmt.Errorf("no session %q", name)
	}

	for _, ws := range sess.Workspaces {
		sway.KillWorkspaceWindows(stashName(name, ws.Slot))
	}

	if err := manager.RunGuarded(func(s *SessionState) error {
		delete(s.Sessions, name)
		return nil
	}); err != nil {
		return err
	}
	log.Infof("Closed session %q", name)
	return nil
}

// Reset clears all session state. Called at sway startup since parked
// workspaces from a previous run no longer exist.
func Reset() {
	if err := getStateManager().RunGuarded(func(s *SessionState) error {
		s.Sessions = map[string]StashedSession{}
		s.Active = map[string]string{}
		return nil
	}); err != nil {
		log.Errorf("Failed to reset session state: %v", err)
	}
}

// CurrentName returns the session name of the currently focused workspace's
// group, or "" if there is none.
func CurrentName() string {
	ws, ok := sway.FocusedWorkspace()
	if !ok {
		return ""
	}
	return groupOf(ws.Num).activeName()
}

func Current() {
	fmt.Print(CurrentName())
}
