package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/backup"
	"github.com/wkozyra95/dotfiles/api/context"
)

func RegisterBackupCmds(rootCmd *cobra.Command) {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Custom backup tool",
	}

	updateCommand := &cobra.Command{
		Use:   "update",
		Short: "Update backups",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := backup.UpdateBackup(ctx); err != nil {
				fmt.Print(err.Error())
			}
		},
	}

	connectCommand := &cobra.Command{
		Use:   "connect",
		Short: "Mount backup volume locally",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := backup.Connect(ctx); err != nil {
				fmt.Print(err.Error())
			}
		},
	}

	restoreCommand := &cobra.Command{
		Use:   "restore",
		Short: "Restore from backups",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := backup.RestoreBackup(ctx); err != nil {
				fmt.Print(err.Error())
			}
		},
	}

	backupCmd.AddCommand(updateCommand)
	backupCmd.AddCommand(restoreCommand)
	backupCmd.AddCommand(connectCommand)

	rootCmd.AddCommand(backupCmd)
}
