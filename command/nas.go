package command

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/nas"
)

func RegisterNasCmds(rootCmd *cobra.Command) {
	nasCmd := &cobra.Command{
		Use:   "nas",
		Short: "home NAS (wake / suspend / rebuild)",
	}

	var noWait bool
	wakeCmd := &cobra.Command{
		Use:   "wake",
		Short: "Send Wake-on-LAN and wait until SSH answers",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if nas.IsUp() {
				log.Info("NAS is already up")
				return
			}
			if err := nas.Wake(); err != nil {
				log.Error(err)
				return
			}
			if noWait {
				return
			}
			log.Info("Magic packet sent, waiting for SSH...")
			if nas.WaitUntilUp(90 * time.Second) {
				log.Info("NAS is up")
			} else {
				log.Error("NAS did not come up within 90s")
			}
		},
	}
	wakeCmd.Flags().BoolVar(&noWait, "no-wait", false, "send the packet and return immediately")

	suspendCmd := &cobra.Command{
		Use:   "suspend",
		Short: "Suspend the NAS to RAM (wake with `mycli nas wake`)",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := nas.Suspend(); err != nil {
				log.Error(err)
			}
		},
	}

	rebuildCmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Build the NAS config locally and switch to it over SSH",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := nas.Rebuild(ctx.FromHome(".dotfiles")); err != nil {
				log.Error(err)
			}
		},
	}

	nasCmd.AddCommand(wakeCmd)
	nasCmd.AddCommand(suspendCmd)
	nasCmd.AddCommand(rebuildCmd)
	rootCmd.AddCommand(nasCmd)
}
