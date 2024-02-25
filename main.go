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

	command.RegisterDriveCmds(rootCmd)
	launcher.RegisterCmds(rootCmd)
	tool.RegisterCmds(rootCmd)
	api.RegisterCmds(rootCmd)
	command.RegisterGitCmds(rootCmd)
	command.RegisterNixCmds(rootCmd)
	command.RegisterCompletionCmds(rootCmd)
	command.RegisterDockerCmds(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
