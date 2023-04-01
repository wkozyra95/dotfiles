package tool

import (
	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
)

func registerUpgradeCommands(rootCmd *cobra.Command) {
	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade system packages",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := ctx.PkgInstaller.UpgradePackages(); err != nil {
				log.Error(err)
			}
		},
	}

	rootCmd.AddCommand(upgradeCmd)
}
