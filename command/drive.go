package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/backup"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/drive"
	"github.com/wkozyra95/dotfiles/api/platform/arch"
)

func RegisterDriveCmds(rootCmd *cobra.Command) {
	driveCmd := &cobra.Command{
		Use:   "drive",
		Short: "drive utils",
	}

	mountCommand := &cobra.Command{
		Use:   "mount",
		Short: "Mount volume",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := drive.Mount(ctx); err != nil {
				fmt.Print(err.Error())
			}
		},
	}

	umountCommand := &cobra.Command{
		Use:   "umount",
		Short: "Umount volume",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := drive.Umount(ctx); err != nil {
				fmt.Print(err.Error())
			}
		},
	}

	updateBackupCommand := &cobra.Command{
		Use:   "backup:update",
		Short: "Update backups",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := backup.UpdateBackup(ctx); err != nil {
				fmt.Print(err.Error())
			}
		},
	}

	restoreBackupCommand := &cobra.Command{
		Use:   "backup:restore",
		Short: "Restore from backups",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := backup.RestoreBackup(ctx); err != nil {
				fmt.Print(err.Error())
			}
		},
	}

	btrfsCleanupCmd := &cobra.Command{
		Use:                   "snapshots:cleanup",
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
		Use:                   "btrfs:restore:root",
		Short:                 "restore root partition",
		Long:                  "",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := arch.RestoreRootSnapshot(); err != nil {
				panic(err)
			}
		},
	}

	driveCmd.AddCommand(mountCommand)
	driveCmd.AddCommand(umountCommand)
	driveCmd.AddCommand(updateBackupCommand)
	driveCmd.AddCommand(restoreBackupCommand)
	driveCmd.AddCommand(btrfsRootRestoreCmd)
	driveCmd.AddCommand(btrfsCleanupCmd)

	rootCmd.AddCommand(driveCmd)
}
