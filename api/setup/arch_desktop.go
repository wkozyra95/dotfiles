package setup

import (
	"fmt"
	"path"
	"regexp"

	a "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

type desktopSetupContext struct {
	username string
}

func (c desktopSetupContext) fromDotfiles(relativePath string) string {
	return path.Join("/home", c.username, ".dotfiles", relativePath)
}

func mainArchStage(ctx desktopSetupContext) a.Object {
	actions := a.List{
		a.ShellCommand("bash", "-c", "echo LANG=en_US.UTF-8 > /etc/locale.conf"),
		a.EnsureText(
			"/etc/locale.gen",
			"en_US.UTF-8 UTF-8",
			regexp.MustCompile(fmt.Sprintf(".*%s.*", regexp.QuoteMeta("en_US.UTF-8 UTF-8"))),
		),
		a.ShellCommand("ln", "-sf", "/usr/share/zoneinfo/Europe/Warsaw", "/etc/localtime"),
		a.ShellCommand("hwclock", "--systohc", "--utc"),
		a.ShellCommand("locale-gen"),
		a.ShellCommand("pacman", "-S", "sudo", "git", "vim", "neovim", "go", "dhcp", "zsh", "openssh"),
		a.Scope("Prompt for hostname:", func() a.Object {
			response := prompt.TextPrompt("/etc/hostname")
			if response == "" {
				return a.Err(fmt.Errorf("Empty value is not allowed"))
			}
			return a.ShellCommand("bash", "-c", fmt.Sprintf("echo \"%s\" > /etc/hostname", response))
		}),
		a.EnsureText(
			"/etc/sudoers",
			"\n%sudo\tALL=(ALL) ALL\n",
			regexp.MustCompile(fmt.Sprintf("\n(.*)%s\n", regexp.QuoteMeta("%sudo\tALL=(ALL) ALL"))),
		),
		a.ShellCommand("groupadd", "sudo"),
		a.ShellCommand("useradd", ctx.username),
		a.ShellCommand("usermod", "-aG", "sudo", ctx.username),
		a.ShellCommand("passwd", ctx.username),
		a.EnsureText(
			"/etc/mkinitcpio.conf",
			"\nHOOKS=(base udev autodetect modconf block encrypt filesystems keyboard btrfs)",
			regexp.MustCompile("\nHOOKS=\\(.*\\)"),
		),
		a.EnsureText(
			"/etc/mkinitcpio.conf",
			"\nFILES=(/root/cryptlvm.keyfile)",
			regexp.MustCompile("\nFILES=\\(.*\\)"),
		),
		a.ShellCommand("mkinitcpio", "-P"),
		a.ShellCommand(
			"chown",
			"-R",
			fmt.Sprintf("%s:%s", ctx.username, ctx.username),
			fmt.Sprintf("/home/%s", ctx.username),
		),
		a.ShellCommand("systemctl", "enable", "NetworkManager"),
		a.ShellCommand("systemctl", "enable", "sshd"),
		a.Func("Select CPU vendor", func() error {
			selected, didSelect := prompt.SelectPrompt(
				"Select CPU vendor",
				[]string{"amd", "intel"},
				func(s string) string { return s },
			)
			if !didSelect {
				return nil
			}
			cmd := exec.Command().WithStdio()
			if selected == "amd" {
				return cmd.Run("pacman", "-S", "amd-ucode")
			} else if selected == "intel" {
				return cmd.Run("pacman", "-S", "intel-ucode")
			}
			return nil
		}),
		a.Func("Select GPU vendor", func() error {
			selected, didSelect := prompt.SelectPrompt(
				"Select GPU vendor",
				[]string{"amd", "intel", "nvidia"},
				func(s string) string { return s },
			)
			if !didSelect {
				return nil
			}
			if selected == "amd" {
				cmd := exec.Command().WithStdio()
				return cmd.Run("pacman", "-S", "vulkan-radeon")
			}
			return nil
		}),
		a.EnsureText(
			fmt.Sprintf("/home/%s/.bash_history", ctx.username),
			"./.dotfiles/bin/mycli tool setup:environment",
			nil,
		),
	}
	return actions
}

func grubInstallStage(ctx desktopSetupContext) a.Object {
	actions := a.List{
		a.WithCondition{
			If: a.PathExists("/sys/firmware/efi"),
			Then: a.ShellCommand(
				"grub-install",
				"--target=x86_64-efi",
				"--efi-directory=/boot/efi",
				"--bootloader-id=GRUB",
			),
			Else: a.ShellCommand("grub-install", "--target=i386-pc", "/dev/sda"),
		},
		a.ShellCommand("grub-mkconfig", "-o", "/boot/grub/grub.cfg"),
	}
	return actions
}

func grubThemeStage(ctx desktopSetupContext) a.Object {
	themePath := "/boot/grub/themes/mytheme"
	return a.List{
		a.ShellCommand("rm", "-rf", themePath),
		a.ShellCommand("cp", "-R", ctx.fromDotfiles("configs/grub-theme"), themePath),
		a.ShellCommand(
			"cp",
			ctx.fromDotfiles("configs/sway/wallpaper.png"),
			path.Join(themePath, "background.png"),
		),
		a.EnsureText(
			"/etc/default/grub",
			fmt.Sprintf("\nGRUB_THEME=\"%s\"\n", path.Join(themePath, "theme.txt")),
			regexp.MustCompile(fmt.Sprintf("\n(.*%s.*)\n", regexp.QuoteMeta("GRUB_THEME"))),
		),
		a.ShellCommand("grub-mkconfig", "-o", "/boot/grub/grub.cfg"),
	}
}

var archDesktopStages = map[string]func(desktopSetupContext) a.Object{
	"main":         mainArchStage,
	"grub-install": grubInstallStage,
	"grub-theme":   grubThemeStage,
}

func ProvisionArchDesktop(stage string) error {
	ctx := desktopSetupContext{
		username: "wojtek",
	}
	if stage == "" {
		return a.RunActions(a.List{
			archDesktopStages["main"](ctx),
			archDesktopStages["grub-install"](ctx),
			archDesktopStages["grub-theme"](ctx),
		})
	} else {
		stageAction, hasStage := archDesktopStages[stage]
		if !hasStage {
			return fmt.Errorf("Stage %s does not exists", stage)
		}
		return a.RunActions(stageAction(ctx))
	}
}
