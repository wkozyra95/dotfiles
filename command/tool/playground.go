package tool

import (
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "docker",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
	},
}

func registerPlaygroundCommands() *cobra.Command {
	rooCmd := &cobra.Command{
		Use:   "playground",
		Short: "playground",
	}

	rooCmd.AddCommand(dockerCmd)
	return rooCmd
}
