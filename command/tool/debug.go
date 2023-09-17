package tool

import (
	"github.com/spf13/cobra"
)

func registerDebugCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "debug",
		Short: "debug",
		Run: func(cmd *cobra.Command, args []string) {
			log.Error("debug")
		},
	}

	return rootCmd
}
