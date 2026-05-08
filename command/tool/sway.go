package tool

import (
	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/sway"
)

func registerSwayListenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sway-listen",
		Short: "Listen to sway workspace events and keep configured workspace pairs in sync across outputs",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := sway.ListenWorkspacePairs(ctx.EnvironmentConfig.SwayWorkspacePairs); err != nil {
				log.Errorf("sway listener exited with error: %v", err)
			}
		},
	}
}
