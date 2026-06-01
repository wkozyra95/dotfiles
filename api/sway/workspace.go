package sway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	swayexec "github.com/wkozyra95/dotfiles/utils/exec"
)

// parkWorkspacePrefix names the hidden, non-numeric placeholder workspaces kept
// on the headless park output. Keeping them non-numeric means they neither
// appear on a visible bar nor hijack `workspace number N` navigation.
const parkWorkspacePrefix = "__park:"

// Workspace is a single entry from `swaymsg -t get_workspaces`.
type Workspace struct {
	Name    string `json:"name"`
	Num     int    `json:"num"`
	Output  string `json:"output"`
	Focused bool   `json:"focused"`
	Visible bool   `json:"visible"`
}

// GetWorkspaces returns the current sway workspaces.
func GetWorkspaces() ([]Workspace, error) {
	var stdout, stderr bytes.Buffer
	err := swayexec.Command().
		WithBufout(&stdout, &stderr).
		Args("swaymsg", "-t", "get_workspaces", "-r").Run()
	if err != nil {
		return nil, err
	}
	var workspaces []Workspace
	if err := json.Unmarshal(stdout.Bytes(), &workspaces); err != nil {
		return nil, err
	}
	return workspaces, nil
}

// FocusedWorkspace returns the currently focused workspace, or false if none
// could be determined.
func FocusedWorkspace() (Workspace, bool) {
	workspaces, err := GetWorkspaces()
	if err != nil {
		return Workspace{}, false
	}
	for _, ws := range workspaces {
		if ws.Focused {
			return ws, true
		}
	}
	return Workspace{}, false
}

// WorkspaceWindowCount walks the sway tree and counts the windows (tiled and
// floating leaves) currently living on the workspace with the given name.
func WorkspaceWindowCount(name string) int {
	tree, err := GetTree()
	if err != nil {
		return 0
	}
	node := FindContainer(tree, func(n TreeNode) bool {
		return n.Type == "workspace" && n.Name == name
	})
	if node == nil {
		return 0
	}
	return countWindows(*node)
}

// WorkspaceWindowCountByNum counts the windows on the workspace with the given
// number, regardless of any label in its name (e.g. "2:smelter").
func WorkspaceWindowCountByNum(num int) int {
	tree, err := GetTree()
	if err != nil {
		return 0
	}
	node := FindContainer(tree, func(n TreeNode) bool {
		return n.Type == "workspace" && n.Num == num
	})
	if node == nil {
		return 0
	}
	return countWindows(*node)
}

// KillWorkspaceWindowsByNum closes every window on the workspace with the given
// number. See killWorkspaceWindows for why con_id targeting is used.
func KillWorkspaceWindowsByNum(num int) {
	killWorkspaceWindows(func(n TreeNode) bool {
		return n.Type == "workspace" && n.Num == num
	})
}

// KillWorkspaceWindows closes every window on the workspace with the exact given
// name. See killWorkspaceWindows for why con_id targeting is used.
func KillWorkspaceWindows(name string) {
	killWorkspaceWindows(func(n TreeNode) bool {
		return n.Type == "workspace" && n.Name == name
	})
}

// killWorkspaceWindows kills each window on the first workspace matching match,
// targeting them by their exact con_id rather than a [workspace="..."] criteria
// — sway matches criteria as a regex, so e.g. "2" would also hit a parked
// "smelter:2", and "smelter:2" would also hit "my-smelter:2".
func killWorkspaceWindows(match func(TreeNode) bool) {
	tree, err := GetTree()
	if err != nil {
		log.Errorf("Failed to read tree: %v", err)
		return
	}
	node := FindContainer(tree, match)
	if node == nil {
		return
	}
	for _, id := range windowConIDs(*node) {
		if err := Command(fmt.Sprintf("[con_id=%d] kill", id)); err != nil {
			log.Errorf("Failed to kill con %d: %v", id, err)
		}
	}
}

// windowConIDs returns the con_ids of the actual windows (leaf containers) under
// node.
func windowConIDs(node TreeNode) []int64 {
	ids := []int64{}
	for _, child := range append(append([]TreeNode{}, node.Nodes...), node.FloatingNodes...) {
		if len(child.Nodes) == 0 && len(child.FloatingNodes) == 0 {
			if child.Type == "con" || child.Type == "floating_con" {
				ids = append(ids, child.ID)
			}
		} else {
			ids = append(ids, windowConIDs(child)...)
		}
	}
	return ids
}

func countWindows(node TreeNode) int {
	count := 0
	for _, child := range append(append([]TreeNode{}, node.Nodes...), node.FloatingNodes...) {
		if len(child.Nodes) == 0 && len(child.FloatingNodes) == 0 {
			// A leaf that is not a layout container is an actual window.
			if child.Type == "con" || child.Type == "floating_con" {
				count++
			}
		} else {
			count += countWindows(child)
		}
	}
	return count
}

// Command runs a single swaymsg command string (e.g. a rename or move). It is a
// package-level counterpart to Listener.Run for callers that are not inside the
// event loop.
func Command(command string) error {
	return swayexec.Command().Args("swaymsg", command).Run()
}

// RenameWorkspace renames the workspace named oldName to newName, preserving all
// of its windows and layout.
func RenameWorkspace(oldName, newName string) error {
	return Command(fmt.Sprintf("rename workspace %q to %q", oldName, newName))
}

// WorkspaceExists reports whether a workspace with the exact given name exists.
func WorkspaceExists(name string) bool {
	workspaces, err := GetWorkspaces()
	if err != nil {
		return false
	}
	for _, ws := range workspaces {
		if ws.Name == name {
			return true
		}
	}
	return false
}

// EnsureParkOutput makes sure a headless output exists to park stashed
// workspaces on (so they stay alive but show on no visible bar) and returns its
// name. The headless output's own resident workspace is renamed to a reserved
// name so it cannot grab a number that a restore wants to reclaim.
func EnsureParkOutput() (string, error) {
	name, findErr := findHeadlessOutput()
	if findErr != nil {
		return "", findErr
	}
	if name == "" {
		if err := Command("create_output"); err != nil {
			return "", err
		}
		name, findErr = findHeadlessOutput()
		if findErr != nil {
			return "", findErr
		}
		if name == "" {
			return "", fmt.Errorf("failed to create headless park output")
		}
		// Park it far off the layout so the cursor never wanders onto it.
		if err := Command(fmt.Sprintf("output %s pos 100000 100000", name)); err != nil {
			log.Errorf("Failed to reposition park output: %v", err)
		}
	}
	SweepParkOutput()
	return name, nil
}

// SweepParkOutput renames any numbered workspace that sway placed (or
// auto-created) on the headless park output to a hidden, non-numeric name. This
// frees the number for navigation/restore and leaves a placeholder behind so
// the park output does not keep spawning numbered replacements.
func SweepParkOutput() {
	name, err := findHeadlessOutput()
	if err != nil || name == "" {
		return
	}
	workspaces, wsErr := GetWorkspaces()
	if wsErr != nil {
		return
	}
	for _, ws := range workspaces {
		if ws.Output == name && ws.Num >= 0 {
			hidden := parkWorkspacePrefix + ws.Name
			if err := RenameWorkspace(ws.Name, hidden); err != nil {
				log.Errorf("Failed to sweep park workspace %s: %v", ws.Name, err)
			}
		}
	}
}

func findHeadlessOutput() (string, error) {
	var stdout, stderr bytes.Buffer
	err := swayexec.Command().
		WithBufout(&stdout, &stderr).
		Args("swaymsg", "-t", "get_outputs", "-r").Run()
	if err != nil {
		return "", err
	}
	var raw []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return "", err
	}
	for _, o := range raw {
		if strings.HasPrefix(o.Name, "HEADLESS-") {
			return o.Name, nil
		}
	}
	return "", nil
}

// MoveWorkspaceToOutput moves the (possibly unfocused) workspace named wsName to
// the given output using workspace criteria, without changing focus. Valid for
// non-empty workspaces.
func MoveWorkspaceToOutput(wsName, output string) error {
	return Command(fmt.Sprintf("[workspace=%q] move workspace to output %q", wsName, output))
}

// FocusWorkspaceNum focuses the numbered workspace n.
func FocusWorkspaceNum(n int) error {
	return Command(fmt.Sprintf("workspace number %d", n))
}

// FocusOutput moves focus to the given output.
func FocusOutput(output string) error {
	return Command(fmt.Sprintf("focus output %q", output))
}

// RealOutputs returns the names of the connected, non-headless outputs.
func RealOutputs() ([]string, error) {
	var stdout, stderr bytes.Buffer
	err := swayexec.Command().
		WithBufout(&stdout, &stderr).
		Args("swaymsg", "-t", "get_outputs", "-r").Run()
	if err != nil {
		return nil, err
	}
	var raw []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return nil, err
	}
	outputs := []string{}
	for _, o := range raw {
		if !strings.HasPrefix(o.Name, "HEADLESS-") {
			outputs = append(outputs, o.Name)
		}
	}
	return outputs, nil
}

// EnsureRealFocus moves focus back to a real output if it is currently on the
// headless park output (which can happen after parking the focused workspace).
// Returns the output focus ended up on, or "" if unknown.
func EnsureRealFocus() string {
	ws, ok := FocusedWorkspace()
	if ok && !strings.HasPrefix(ws.Output, "HEADLESS-") {
		return ws.Output
	}
	outputs, err := RealOutputs()
	if err != nil || len(outputs) == 0 {
		return ws.Output
	}
	if err := FocusOutput(outputs[0]); err != nil {
		log.Errorf("Failed to focus real output %s: %v", outputs[0], err)
		return ws.Output
	}
	return outputs[0]
}
