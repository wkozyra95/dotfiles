package tool

import (
	"os/user"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/setup"
)

func registerSetupCommands(rootCmd *cobra.Command) {
	var setupEnvReinstall bool
	setupEnvCmd := &cobra.Command{
		Use:   "setup:environment",
		Short: "Install tools + prepare config files",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := setup.SetupEnvironment(ctx, setup.SetupEnvironmentOptions{
				Reinstall: setupEnvReinstall,
			}); err != nil {
				log.Error(err)
			}
		},
	}

	setupArchInstallerCmd := &cobra.Command{
		Use:   "setup:arch:usb",
		Short: "prepare Arch installer",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := setup.ProvisionUsbArchInstaller(ctx); err != nil {
				log.Error(err)
			}
		},
	}

	setupArchChrootCmd := &cobra.Command{
		Use:   "setup:arch:chroot",
		Short: "prepare chrooted environment",
		Run: func(cmd *cobra.Command, args []string) {
			user, userErr := user.Current()
			if userErr != nil {
				log.Error(userErr)
				return
			}
			if user.Username != "root" {
				log.Error("You need to be root to run installer")
				return
			}
			if err := setup.ProvisionArchChroot(); err != nil {
				log.Error(err)
			}
		},
	}

	setupArchCompanionChrootCmd := &cobra.Command{
		Use:   "setup:arch:companion_chroot",
		Short: "prepare chrooted environment for companion system",
		Run: func(cmd *cobra.Command, args []string) {
			user, userErr := user.Current()
			if userErr != nil {
				log.Error(userErr)
				return
			}
			if user.Username != "root" {
				log.Error("You need to be root to run installer")
				return
			}
			if err := setup.ProvisionArchChrootForCompanionSystem(); err != nil {
				log.Error(err)
			}
		},
	}

	connectExistingArchChrootCmd := &cobra.Command{
		Use:   "connect:arch:chroot",
		Short: "connect to existing chrooted environment",
		Run: func(cmd *cobra.Command, args []string) {
			user, userErr := user.Current()
			if userErr != nil {
				log.Error(userErr)
				return
			}
			if user.Username != "root" {
				log.Error("You need to be root to run installer")
				return
			}
			if err := setup.ConnectToExistingChrootedEnvironment(); err != nil {
				log.Error(err)
			}
		},
	}

	var desktopArchStage string
	setupDesktopArchCmd := &cobra.Command{
		Use:   "setup:arch:desktop",
		Short: "setup desktop arch",
		Run: func(cmd *cobra.Command, args []string) {
			user, userErr := user.Current()
			if userErr != nil {
				log.Error(userErr)
				return
			}
			if user.Username != "root" {
				log.Error("You need to be root to run installer")
				return
			}
			if err := setup.ProvisionArchDesktop(desktopArchStage); err != nil {
				log.Error(err)
			}
		},
	}

	setupInDockerCmd := &cobra.Command{
		Use:   "setup:environment:docker",
		Short: "command that should be run inside dockerfile that provisions the system",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := setup.SetupUbuntuInDocker(ctx, setup.SetupEnvironmentOptions{Reinstall: true}); err != nil {
				log.Errorln(err.Error())
			}
		},
	}

	setupDesktopArchCmd.PersistentFlags().StringVar(
		&desktopArchStage, "stage", "",
		"setup only specified stage",
	)

	setupEnvCmd.PersistentFlags().BoolVar(
		&setupEnvReinstall, "reinstall", false,
		"reinstall packages even if they are already installed",
	)

	rootCmd.AddCommand(setupEnvCmd)
	rootCmd.AddCommand(setupArchInstallerCmd)
	rootCmd.AddCommand(setupArchChrootCmd)
	rootCmd.AddCommand(setupArchCompanionChrootCmd)
	rootCmd.AddCommand(setupDesktopArchCmd)
	rootCmd.AddCommand(connectExistingArchChrootCmd)
	rootCmd.AddCommand(setupInDockerCmd)
}
