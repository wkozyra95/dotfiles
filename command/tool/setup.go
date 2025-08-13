package tool

import (
	"os/user"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/setup"
)

func registerSetupCommands(rootCmd *cobra.Command) {
	var setupEnvOptions setup.SetupEnvironmentOptions
	setupEnvCmd := &cobra.Command{
		Use:   "setup:environment",
		Short: "Install tools + prepare config files",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := setup.SetupEnvironment(ctx, setupEnvOptions); err != nil {
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

	setupNixInstallerCmd := &cobra.Command{
		Use:   "setup:nix:usb",
		Short: "prepare Nix installer",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := setup.ProvisionUsbNixInstaller(ctx); err != nil {
				log.Error(err)
			}
		},
	}

	installNixOSCmd := &cobra.Command{
		Use:   "install:nixos",
		Short: "install NixOS",
		Run: func(cmd *cobra.Command, args []string) {
			user, userErr := user.Current()
			if userErr != nil {
				log.Error(userErr)
				return
			}
			if user.Username != "root" {
				log.Errorf("This command should only be ran from installer medium. (Username=wojtek expected, found: %v)", user.Username)
				return
			}
			if err := setup.InstallNixOS(); err != nil {
				log.Error(err)
			}
		},
	}

	setupInDockerCmd := &cobra.Command{
		Use:   "setup:environment:docker",
		Short: "command that should be run inside dockerfile that provisions the system",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContextForEnvironment("docker")
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
		&setupEnvOptions.Reinstall, "reinstall", false,
		"reinstall packages even if they are already installed",
	)
	setupEnvCmd.PersistentFlags().BoolVar(
		&setupEnvOptions.DryRun, "dry-run", false,
		"only print actions",
	)

	rootCmd.AddCommand(setupEnvCmd)
	rootCmd.AddCommand(setupArchInstallerCmd)
	rootCmd.AddCommand(setupArchChrootCmd)
	rootCmd.AddCommand(setupArchCompanionChrootCmd)
	rootCmd.AddCommand(setupDesktopArchCmd)
	rootCmd.AddCommand(setupNixInstallerCmd)
	rootCmd.AddCommand(installNixOSCmd)
	rootCmd.AddCommand(connectExistingArchChrootCmd)
	rootCmd.AddCommand(setupInDockerCmd)
}
