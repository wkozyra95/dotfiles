package main

import (
	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/command"
	"github.com/wkozyra95/dotfiles/command/api"
	"github.com/wkozyra95/dotfiles/command/launcher"
	"github.com/wkozyra95/dotfiles/command/tool"
	"github.com/wkozyra95/dotfiles/logger"
)

var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "Set of system tools",
}

var log = logger.NamedLogger("main")

func main() {
	log.Debug("main()")

	command.RegisterBackupCmds(rootCmd)
	launcher.RegisterCmds(rootCmd)
	tool.RegisterCmds(rootCmd)
	api.RegisterCmds(rootCmd)
	command.RegisterBtrfsCmds(rootCmd)
	command.RegisterGitCmds(rootCmd)
	command.RegisterCompletionCmds(rootCmd)

	rootCmd.Execute()
}
