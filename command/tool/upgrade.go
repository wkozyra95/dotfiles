package tool

import (
	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/platform"
)

func registerUpgradeCommands(rootCmd *cobra.Command) {
	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade system packages",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			pkgInstaller, pkgInstallerErr := platform.GetPackageManager(ctx)
			if pkgInstallerErr != nil {
				panic(pkgInstallerErr)
			}
			if err := pkgInstaller.UpgradePackages(); err != nil {
				log.Error(err)
			}
		},
	}

	rootCmd.AddCommand(upgradeCmd)
}
