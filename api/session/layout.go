package session

import (
	"fmt"
	"strings"
	"time"

	"github.com/wkozyra95/dotfiles/api/sway"
)

const (
	// layoutMark is the temporary sway mark used as the target of container moves.
	layoutMark = "_session_layout"

	windowWaitTimeout  = 10 * time.Second
	windowPollInterval = 100 * time.Millisecond
)

// launchedWindow is a freshly started terminal process awaiting layout.
type launchedWindow struct {
	pid int
	// exited is closed once the process has exited (and been reaped).
	exited <-chan struct{}
}

func (w launchedWindow) hasExited() bool {
	select {
	case <-w.exited:
		return true
	default:
		return false
	}
}

// arrangeColumns lays out freshly launched terminals, given per slot: slot 0 in
// a left column and slot 1 in a right one, each stacked vertically in launch
// order, or a single row if only one slot has windows. Focus moves to the first
// window of the left column, as long as the user is still on workspace ws.
//
// Windows are arranged after launch rather than placed as they start, because
// they map in an arbitrary order. A window that fails to appear in time, or
// whose process exits before it is arranged, is left out.
func arrangeColumns(launchedBySlot [2][]launchedWindow, ws int) {
	launched := append(append([]launchedWindow{}, launchedBySlot[0]...), launchedBySlot[1]...)
	if len(launched) < 2 {
		return
	}
	windows := waitForWindows(launched)
	if len(windows) < 2 {
		return
	}
	conByPID := map[int]int64{}
	for _, w := range windows {
		conByPID[w.PID] = w.ID
	}
	columns := [][]int64{}
	for _, slotWindows := range launchedBySlot {
		column := []int64{}
		for _, w := range slotWindows {
			if con, ok := conByPID[w.pid]; ok {
				column = append(column, con)
			}
		}
		if len(column) > 0 {
			columns = append(columns, column)
		}
	}
	// Sent as one batch so sway applies it in a single transaction, without
	// rendering the intermediate layouts.
	if err := sway.Command(strings.Join(columnLayoutCommands(windows[0].ID, columns), "; ")); err != nil {
		log.Errorf("Failed to arrange session windows: %v", err)
	}
	if focused, ok := sway.FocusedWorkspace(); ok && focused.Num == ws {
		if err := sway.Command(fmt.Sprintf("[con_id=%d] focus", columns[0][0])); err != nil {
			log.Errorf("Failed to focus session window: %v", err)
		}
	}
}

// waitForWindows polls until every still-running process has a window or the
// timeout expires, and returns the windows found so far, in tree order. A
// process that exits (e.g. a NoShellFallback task that finished quickly) stops
// being waited for, and its window, if still mapped, is not returned.
func waitForWindows(launched []launchedWindow) []sway.TreeNode {
	deadline := time.Now().Add(windowWaitTimeout)
	for {
		pids := []int{}
		for _, w := range launched {
			if !w.hasExited() {
				pids = append(pids, w.pid)
			}
		}
		windows, err := sway.WindowsByPID(pids)
		if err != nil {
			log.Errorf("Failed to read tree: %v", err)
		}
		if len(windows) == len(pids) {
			return windows
		}
		if time.Now().After(deadline) {
			log.Errorf("Only %d of %d session windows appeared in time", len(windows), len(pids))
			return windows
		}
		time.Sleep(windowPollInterval)
	}
}

// columnLayoutCommands returns the sway commands that turn a flat row of windows
// into the given columns (one column means a plain row, not a stack). first is
// the window currently leftmost in the row.
func columnLayoutCommands(first int64, columns [][]int64) []string {
	anchor := columns[0][0]
	cmds := []string{fmt.Sprintf("[con_id=%d] layout splith", anchor)}
	if first != anchor {
		cmds = append(cmds, fmt.Sprintf("[con_id=%d] swap container with con_id %d", anchor, first))
	}
	if len(columns) == 1 {
		cmds = append(cmds, chainCommands(columns[0])...)
	} else {
		// Place the right column's head next to the left's while the row is still
		// flat, then wrap each column's head in a vertical split and pull the
		// rest of that column in behind it.
		cmds = append(cmds, chainCommands([]int64{anchor, columns[1][0]})...)
		for _, column := range columns {
			if len(column) > 1 {
				cmds = append(cmds, fmt.Sprintf("[con_id=%d] splitv", column[0]))
				cmds = append(cmds, chainCommands(column)...)
			}
		}
	}
	return append(cmds, "unmark "+layoutMark)
}

// chainCommands moves each window directly after the previous one in the list.
func chainCommands(windows []int64) []string {
	cmds := []string{}
	for i := 1; i < len(windows); i++ {
		cmds = append(cmds,
			fmt.Sprintf("[con_id=%d] mark --add %s", windows[i-1], layoutMark),
			fmt.Sprintf("[con_id=%d] move container to mark %s", windows[i], layoutMark),
		)
	}
	return cmds
}
