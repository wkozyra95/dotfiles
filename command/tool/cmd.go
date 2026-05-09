package tool

import (
	"github.com/spf13/cobra"
)

// RegisterCmds ...
func RegisterCmds(rootCmd *cobra.Command) {
	toolCmd := &cobra.Command{
		Use:   "tool",
		Short: "system tools",
	}

	registerSetupCommands(toolCmd)
	registerUpgradeCommands(toolCmd)
	registerReleaseCommands(toolCmd)
	toolCmd.AddCommand(registerWifiCommands())
	toolCmd.AddCommand(registerPlaygroundCommands())
	toolCmd.AddCommand(registerDebugCommand())
	rootCmd.AddCommand(toolCmd)
}
