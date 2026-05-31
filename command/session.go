package command

import (
	"errors"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/session"
	"github.com/wkozyra95/dotfiles/utils/notify"
	"github.com/wkozyra95/dotfiles/utils/term"
)

// withSessionLog runs a session action with stdout/stderr redirected to a
// logfile (these commands are launched detached from a terminal via sway
// `exec`, so their log output would otherwise be lost) and surfaces any error
// as a desktop notification. A user cancellation (ErrCanceled) is a silent
// no-op.
func withSessionLog(fn func() error) {
	run := func() {
		if err := fn(); err != nil && !errors.Is(err, session.ErrCanceled) {
			notify.Notify("Session", err.Error())
		}
	}
	logfile := "/tmp/mycli/session.log"
	if err := os.MkdirAll(path.Dir(logfile), os.ModePerm); err != nil {
		run()
		return
	}
	redirects, redirectErr := term.RedirectStdioToFile(logfile)
	if redirectErr != nil {
		run()
		return
	}
	defer redirects.Cleanup()
	run()
}

// RegisterSessionCmds wires the `session` command and its subcommands into the
// root command. All logic lives in api/session.
func RegisterSessionCmds(rootCmd *cobra.Command) {
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "sway session manager",
	}

	newCmd := &cobra.Command{
		Use:   "new",
		Short: "stash the current unit (if occupied) and launch a template into it",
		Run: func(cmd *cobra.Command, args []string) {
			withSessionLog(func() error { return session.New(context.CreateContext()) })
		},
	}

	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "switch to a stashed session",
		Run: func(cmd *cobra.Command, args []string) {
			withSessionLog(session.Switch)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list the active and stashed sessions",
		Run: func(cmd *cobra.Command, args []string) {
			if err := session.List(); err != nil {
				log.Error(err)
			}
		},
	}

	closeCmd := &cobra.Command{
		Use:   "close [name]",
		Short: "close a stashed session (kill its windows)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			withSessionLog(func() error { return session.Close(name) })
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "print the session name of the focused workspace (for the bar)",
		Run: func(cmd *cobra.Command, args []string) {
			// No log redirect: stdout must stay clean for status_command.
			session.Current()
		},
	}

	sessionCmd.AddCommand(newCmd)
	sessionCmd.AddCommand(switchCmd)
	sessionCmd.AddCommand(listCmd)
	sessionCmd.AddCommand(closeCmd)
	sessionCmd.AddCommand(currentCmd)
	rootCmd.AddCommand(sessionCmd)
}
