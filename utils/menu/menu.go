// Package menu wraps bemenu to present graphical pickers (and free-text
// prompts) from sway keybindings without opening a terminal window.
package menu

import (
	"os/exec"
	"strings"

	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("menu")

func run(prompt string, items []string) (string, bool) {
	cmd := exec.Command(
		"bemenu",
		"-i",       // case-insensitive matching
		"-l", "15", // show up to 15 entries as a vertical list
		"-p", prompt,
	)
	if len(items) > 0 {
		cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))
	} else {
		cmd.Stdin = strings.NewReader("")
	}
	out, err := cmd.Output()
	if err != nil {
		// non-zero exit means the user pressed Escape / cancelled
		log.Debugf("bemenu cancelled or failed: %v", err)
		return "", false
	}
	result := strings.TrimRight(string(out), "\n")
	if result == "" {
		return "", false
	}
	return result, true
}

// Select shows the given items in bemenu and returns the chosen one. The second
// return value is false when the user cancelled.
func Select(prompt string, items []string) (string, bool) {
	return run(prompt, items)
}

// Prompt shows an empty bemenu and returns whatever text the user typed (used
// for free-text input such as a new session name).
func Prompt(prompt string) (string, bool) {
	return run(prompt, nil)
}
