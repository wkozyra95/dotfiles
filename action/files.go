package action

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
	"github.com/wkozyra95/dotfiles/utils/http"
)

func EnsureSymlink(source string, destination string) Object {
	return SimpleAction{
		Run: func() error {
			return file.EnsureSymlink(source, destination)
		},
		Label: fmt.Sprintf("Symlink(%s -> %s)", destination, source),
	}
}

func EnsureText(path string, text string, rg *regexp.Regexp) Object {
	split := strings.Split(text, "\n")
	if len(split) > 2 {
		split = split[0:2]
	}
	joined := strings.Join(split, "\n# ")
	label := fmt.Sprintf("ensure text %s\n# %s", path, joined)

	return SimpleAction{
		Run: func() error {
			return file.EnsureTextWithRegexp(path, text, rg)
		},
		Label: label,
	}
}

func ShellCommand(args ...string) Object {
	label := fmt.Sprintf("Shell(%s)", strings.Join(args, " "))
	return SimpleAction{
		Run: func() error {
			return exec.Command().WithStdio().Run(args[0], args[1:]...)
		},
		Label: label,
	}
}

func Execute(cmd *exec.Cmd, args ...string) Object {
	label := fmt.Sprintf("Shell(%s)", strings.Join(args, " "))
	return SimpleAction{
		Run: func() error {
			return cmd.WithStdio().Run(args[0], args[1:]...)
		},
		Label: label,
	}
}

func PathExists(path string) Condition {
	return SimpleCondition{
		Check: func() (bool, error) {
			return file.Exists(path), nil
		},
		Label: fmt.Sprintf("PathExists(%s)", path),
	}
}

func CommandExists(cmd string) Condition {
	return SimpleCondition{
		Check: func() (bool, error) {
			return exec.CommandExists(cmd), nil
		},
		Label: fmt.Sprintf("CommandExists(%s)", cmd),
	}
}

func DownloadFile(url string, path string) Object {
	return SimpleAction{
		Run: func() error {
			return http.DownloadFile(url, path)
		},
		Label: fmt.Sprintf("DownloadFile(%s -> %s)", url, path),
	}
}
