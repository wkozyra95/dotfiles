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

func run(prompt string, items []string, extraArgs ...string) (string, bool) {
	// fuzzel --dmenu reads newline-separated entries from stdin and prints the
	// selection; with no matching entry it prints the raw input (used by Prompt
	// for free-text such as a new session name). Matching is case-insensitive.
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
	result := strings.TrimRight(string(out), "\n")
	if result == "" {
		return "", false
	}
	return result, true
}

// Select shows the given items in fuzzel and returns the chosen one. The second
// return value is false when the user cancelled.
func Select(prompt string, items []string) (string, bool) {
	return run(prompt, items)
}

// Prompt shows fuzzel as a single-line text input (no result list) and returns
// whatever text the user typed (used for free-text input such as a new session
// name). --lines=0 collapses the list area so it doesn't render tall and empty.
func Prompt(prompt string) (string, bool) {
	return run(prompt, nil, "--lines=0")
}
