package sway

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/fn"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	log         = logger.NamedLogger("sway")
	workspaceRg = regexp.MustCompile(`workspace\d+`)
)

// OpenTerminal opens terminal in the same directory as a
// shell in currently opened window. If current window does
// not run a shell, it should fallback to home directory.
func OpenTerminal(ctx context.Context) error {
	err := maybeOpenTerminalInTheSameDirectory()
	if err != nil {
		log.Error(err.Error())
		return api.AlacrittyCall(
			api.AlacrittyConfig{Command: "zsh", Args: []string{}, Cwd: ctx.Homedir, ShouldRetry: false},
		)
	}
	return nil
}

func maybeOpenTerminalInTheSameDirectory() error {
	tree, err := GetTree()
	if err != nil {
		return err
	}
	node := FindContainer(tree, func(tn TreeNode) bool {
		return tn.Focused && (tn.AppID == "Alacritty" || workspaceRg.Match([]byte(tn.AppID))) && tn.Visible &&
			tn.Type == "con"
	})
	if node == nil {
		return errors.New("node not found")
	}
	childPid := getPidOfLastDescendantRunningZsh(node.PID)
	if childPid == 0 {
		return errors.New("now child processes run zsh")
	}

	destination, readErr := os.Readlink(fmt.Sprintf("/proc/%d/cwd", childPid))
	if readErr != nil {
		return readErr
	}

	return api.AlacrittyCall(
		api.AlacrittyConfig{Command: "zsh", Cwd: destination, ShouldRetry: false},
	)
}

// getPidOfLastDescendantRunningZsh resolves a process that is likely to
// be a current interactive shell.
//
//   - Resolves all children of a process by applying "pgrep -P" until
//     all descendants are resolved
//   - Iterate over children in a descending order (to make sure oldest
//     shell is returned first). This solution relies on process PIDs
//     increasing monotonically.
//   - Match processes that have "/proc/${pid}/cmdline" equal to "zsh"
//     or "/bin/zsh--login"
//   - TODO: instead of checking by name, search for process that has
//     TTY attached.
func getPidOfLastDescendantRunningZsh(pid int) int {
	pidsMap := map[int]struct{}{pid: {}}
	for {
		keys := maps.Keys(pidsMap)
		childrenPids := pgrepP(keys)
		addedNewPid := false
		for _, childPid := range childrenPids {
			_, hasKey := pidsMap[childPid]
			if !hasKey {
				pidsMap[childPid] = struct{}{}
				addedNewPid = true
			}
		}
		if !addedNewPid {
			break
		}
	}
	pids := maps.Keys(pidsMap)
	slices.Sort(pids)
	for i := 0; i < len(pids); i++ {
		pid := pids[len(pids)-i-1]
		cmdline, readErr := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		if readErr != nil {
			log.Error(readErr.Error())
			continue
		}
		// /proc/pid/cmdline is using byte 0 a separator, so we need to remove it
		trimed := strings.Replace(string(cmdline), string([]byte{0}), "", -1)
		if trimed == "zsh" || trimed == "zsh--login" {
			return pid
		}
	}
	return 0
}

func pgrepP(ppids []int) []int {
	var stdout, stderr bytes.Buffer
	err := exec.Command().
		WithBufout(&stdout, &stderr).
		Args("pgrep", "-P", strings.Join(fn.Map(ppids, strconv.Itoa), ",")).Run()
	if err != nil {
		return []int{}
	}
	split := strings.Split(stdout.String(), "\n")
	parsed := fn.Map(split, func(i string) int {
		value, err := strconv.Atoi(i)
		if err != nil {
			return 0
		}
		return value
	})
	return fn.Filter(parsed, func(i int) bool { return i != 0 })
}
