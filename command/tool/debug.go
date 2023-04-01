package tool

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/platform/arch"
)

func registerDebugCommand() *cobra.Command {
	rooCmd := &cobra.Command{
		Use:   "debug",
		Short: "debug",
		Run: func(cmd *cobra.Command, args []string) {
			spew.Dump(arch.GetSnapshots())
		},
	}

	return rooCmd
}
