package action

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
	"github.com/wkozyra95/dotfiles/utils/http"
)

type ensureSymlinkArgs struct {
	source      string
	destination string
}

var ensureSymlink = SimpleActionBuilder[ensureSymlinkArgs]{
	CreateRun: func(args ensureSymlinkArgs) func() error {
		return func() error {
			return file.EnsureSymlink(args.source, args.destination)
		}
	},
	String: func(args ensureSymlinkArgs) string {
		return fmt.Sprintf("Symlink(%s -> %s)", args.destination, args.source)
	},
}.Init()

func EnsureSymlink(source string, destination string) Object {
	return ensureSymlink(ensureSymlinkArgs{source: source, destination: destination})
}

type ensureTextArgs struct {
	Path   string
	Text   string
	Regexp *regexp.Regexp
}

var ensureText = SimpleActionBuilder[ensureTextArgs]{
	CreateRun: func(args ensureTextArgs) func() error {
		return func() error {
			return file.EnsureTextWithRegexp(args.Path, args.Text, args.Regexp)
		}
	},
	String: func(args ensureTextArgs) string {
		text := strings.Split(args.Text, "\n")
		if len(text) > 3 {
			text = text[0:2]
		}
		joined := strings.Join(text, "\n# ")
		return fmt.Sprintf("ensure text %s\n# %s", args.Path, joined)
	},
}.Init()

func EnsureText(path string, text string, rg *regexp.Regexp) Object {
	return ensureText(ensureTextArgs{
		Path:   path,
		Text:   text,
		Regexp: rg,
	})
}

type commandActionArgs struct {
	args []string
	cmd  *exec.Cmd
}

var commandAction = SimpleActionBuilder[commandActionArgs]{
	CreateRun: func(c commandActionArgs) func() error {
		return func() error {
			return c.cmd.Run(c.args[0], c.args[1:]...)
		}
	},
	String: func(c commandActionArgs) string {
		return strings.Join(c.args, " ")
	},
}.Init()

func ShellCommand(args ...string) Object {
	return commandAction(commandActionArgs{
		args: append([]string{}, args...),
		cmd:  exec.Command().WithStdio(),
	})
}

func Execute(cmd *exec.Cmd, args ...string) Object {
	return commandAction(commandActionArgs{
		args: append([]string{}, args...),
		cmd:  cmd,
	})
}

var PathExists = SimpleConditionBuilder[string]{
	CreateCondition: func(arg string) func() (bool, error) {
		return func() (bool, error) {
			return file.Exists(arg), nil
		}
	},
	String: func(s string) string {
		return fmt.Sprintf("PathExists(%s)", s)
	},
}.Init()

var CommandExists = SimpleConditionBuilder[string]{
	CreateCondition: func(arg string) func() (bool, error) {
		return func() (bool, error) {
			return exec.CommandExists(arg), nil
		}
	},
	String: func(s string) string {
		return fmt.Sprintf("CommandExists(%s)", s)
	},
}.Init()

type downloadFileArgs struct {
	path string
	url  string
}

var downloadFile = SimpleActionBuilder[downloadFileArgs]{
	CreateRun: func(dfa downloadFileArgs) func() error {
		return func() error {
			return http.DownloadFile(dfa.url, dfa.path)
		}
	},
	String: func(dfa downloadFileArgs) string {
		return fmt.Sprintf("DownloadFile(%s -> %s)", dfa.url, dfa.path)
	},
}.Init()

func DownloadFile(url string, path string) Object {
	return downloadFile(downloadFileArgs{
		path: path,
		url:  url,
	})
}
