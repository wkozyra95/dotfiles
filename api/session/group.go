package session

import (
	"strconv"

	"github.com/wkozyra95/dotfiles/api/sway"
)

// workspaceGroup is the set of workspaces a session occupies: a pair (2+6 or
// 3+7) or a standalone workspace (partner == primary).
type workspaceGroup struct {
	primary int
	partner int
}

// groupOf returns the group containing workspace num. Pairs are 2+6 and 3+7;
// everything else is standalone.
func groupOf(num int) workspaceGroup {
	switch num {
	case 2, 6:
		return workspaceGroup{primary: 2, partner: 6}
	case 3, 7:
		return workspaceGroup{primary: 3, partner: 7}
	default:
		return workspaceGroup{primary: num, partner: num}
	}
}

// standalone reports whether the group is a single workspace (no partner).
func (g workspaceGroup) standalone() bool { return g.primary == g.partner }

// workspaces returns the group's distinct workspace numbers.
func (g workspaceGroup) workspaces() []int {
	if g.standalone() {
		return []int{g.primary}
	}
	return []int{g.primary, g.partner}
}

// key identifies the group (by its primary workspace) in the Active map.
func (g workspaceGroup) key() string { return strconv.Itoa(g.primary) }

// windowCount sums the windows currently living on the group's workspaces.
func (g workspaceGroup) windowCount() int {
	total := 0
	for _, n := range g.workspaces() {
		total += sway.WorkspaceWindowCountByNum(n)
	}
	return total
}

// outputs decides which output each of the group's workspaces should live on,
// keeping the focused workspace on its current output (originOut) and placing
// the other on the second monitor. This way restoring/launching while focused
// on the partner (e.g. ws6) still puts the primary (ws2) on the correct screen.
func (g workspaceGroup) outputs(focusedNum int, originOut string) (primaryOut, partnerOut string) {
	if originOut == "" {
		if ws, ok := sway.FocusedWorkspace(); ok {
			originOut = ws.Output
		}
	}
	primaryOut, partnerOut = originOut, otherOutput(originOut)
	if focusedNum == g.partner && !g.standalone() {
		primaryOut, partnerOut = otherOutput(originOut), originOut
	}
	return primaryOut, partnerOut
}

// slotOf returns 0 if num is a primary workspace, 1 if it is a partner.
func slotOf(num int) int {
	if num == groupOf(num).primary {
		return 0
	}
	return 1
}

// otherOutput returns a real output different from origin (for placing a pair's
// partner on the second monitor), falling back to origin if there is only one.
func otherOutput(origin string) string {
	outputs, err := sway.RealOutputs()
	if err != nil {
		return origin
	}
	for _, o := range outputs {
		if o != origin {
			return o
		}
	}
	return origin
}
