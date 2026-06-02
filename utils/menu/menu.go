// Package menu wraps fuzzel to present centered graphical pickers (and
// free-text prompts) from sway keybindings without opening a terminal window.
// Styling lives in ~/.config/fuzzel/fuzzel.ini.
package menu

import (
	"os/exec"
	"strings"

	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("menu")

// run shows fuzzel and returns the selection. The second return is false on
// cancel; since fuzzel treats Enter on empty input as a cancel (non-zero exit),
// an empty submission is indistinguishable from cancel and also returns
// ("", false) — there is no way to submit an empty string.
func run(prompt string, items []string, extraArgs ...string) (string, bool) {
	args := append([]string{"--dmenu", "--prompt", prompt + " "}, extraArgs...)
	cmd := exec.Command("fuzzel", args...)
	if len(items) > 0 {
		cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))
	} else {
		cmd.Stdin = strings.NewReader("")
	}
	out, err := cmd.Output()
	if err != nil {
		// non-zero exit means the user pressed Escape / cancelled
		log.Debugf("fuzzel cancelled or failed: %v", err)
		return "", false
	}
	return strings.TrimRight(string(out), "\n"), true
}

// Select picks one of items; the second return is false when nothing was chosen.
func Select(prompt string, items []string) (string, bool) {
	result, ok := run(prompt, items)
	if !ok || result == "" {
		return "", false
	}
	return result, true
}

// Prompt reads a line of free text, returning "" on cancel. An empty line
// cannot be submitted (Enter on empty input cancels), so "" always means cancel.
func Prompt(prompt string) string {
	result, _ := run(prompt, nil, "--lines=0")
	return result
}
