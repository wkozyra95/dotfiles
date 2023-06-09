package ubuntu

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

var _ api.PackageInstaller = Apt{}

type aptPackage []string

func (p aptPackage) Install() error {
	user, userErr := user.Current()
	if userErr != nil {
		return userErr
	}
	for _, pkg := range p {
		cmd := exec.Command().WithStdio()
		if user.Name != "root" {
			cmd = cmd.WithSudo()
		}
		installErr := cmd.Run("apt-get", "install", "-y", pkg)
		if installErr != nil && !prompt.ConfirmPrompt("Install failed, do you want to continue?") {
			return installErr
		}
	}
	return nil
}

func (a Apt) UpgradePackages() error {
	panic("Upgrading all packages is not supported")
}

func (p aptPackage) String() string {
	return fmt.Sprintf("apt-get install -y %s", strings.Join(p, " "))
}

type Apt struct{}

func (a Apt) DevelopmentTools() api.Package {
	return aptPackage{"clang", "cmake", "ninja-build"}
}

func (a Apt) ShellTools() api.Package {
	return aptPackage{
		"build-essential",

		"zsh",
		"vim",
		"neovim",
		"htop",
		"ripgrep",
		"silversearcher-ag",
		"fzf",
		//"diff-so-fancy",
		"git",
		"git-crypt",
		//"fd",
		"unzip",
		"python3-pip",
		"rsync",
		"ranger",
		//"clojure-lsp-bin",
		"jq",

		"ssh",
		"btop",
		"curl",
		"wget",

		// for neovim form source
		"gettext", "libtool-bin", "g++", "pkg-config",
	}
}

func (a Apt) Desktop() api.Package {
	panic("unsupported package group; desktop")
}

func (a Apt) CustomPackageList(pkgs []string) api.Package {
	return aptPackage(pkgs)
}

func (a Apt) EnsurePackagerAction(homedir string) action.Object {
	user, _ := user.Current()
	if user != nil && user.Name != "root" {
		return action.ShellCommand("sudo", "apt-get", "update", "-y")
	} else {
		return action.ShellCommand("apt-get", "update", "-y")
	}
}
