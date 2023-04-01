package command

import (
	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/platform/arch"
)

func RegisterBtrfsCmds(rootCmd *cobra.Command) {
	btrfsCmd := &cobra.Command{
		Use:   "btrfs",
		Short: "btrfs tooling",
		Long:  "",
	}

	btrfsCleanupCmd := &cobra.Command{
		Use:                   "cleanup",
		Short:                 "cleanup snapshots",
		Long:                  "",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := arch.CleanupSnapshots(); err != nil {
				panic(err)
			}
		},
	}

	btrfsRootRestoreCmd := &cobra.Command{
		Use:                   "restore:root",
		Short:                 "restore root partition",
		Long:                  "",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := arch.RestoreRootSnapshot(); err != nil {
				panic(err)
			}
		},
	}

	btrfsCmd.AddCommand(btrfsRootRestoreCmd)
	btrfsCmd.AddCommand(btrfsCleanupCmd)

	rootCmd.AddCommand(btrfsCmd)
}
