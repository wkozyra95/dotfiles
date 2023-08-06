package arch

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

var _ api.PackageInstaller = Yay{}

type yayPackage []string

func (p yayPackage) Install() error {
	for _, pkg := range p {
		var stdout bytes.Buffer
		if err := exec.Command().WithBufout(&stdout, &bytes.Buffer{}).Run("yay", "-Qi", pkg); err == nil {
			continue
		}
		installErr := exec.Command().WithStdio().Run("yay", "-S", pkg)
		if installErr != nil && !prompt.ConfirmPrompt("Install failed, do you want to continue?") {
			return installErr
		}
	}
	return nil
}

func (p yayPackage) String() string {
	return fmt.Sprintf("yay -S %s", strings.Join(p, " "))
}

type Yay struct{}

func (y Yay) DevelopmentTools() api.Package {
	return yayPackage{"clang", "cmake", "go", "ninja"}
}

func (y Yay) ShellTools() api.Package {
	return yayPackage{
		"zsh",
		"vim",
		"neovim",
		"htop",
		"ripgrep",
		"the_silver_searcher",
		"fzf",
		"diff-so-fancy",
		"git",
		"git-crypt",
		"fd",
		"unzip",
		"python-pip",
		"rsync",
		"ranger",
		"clojure-lsp-bin",
		"jq",
		"python-pipx",
	}
}

func (y Yay) Desktop() api.Package {
	return yayPackage{
		"gnu-free-fonts",
		"adobe-source-code-pro-fonts",
		"ttf-nerd-fonts-symbols-mono",

		"pipewire",
		"pipewire-pulse",
		"alsa-utils",
		"pamixer",
		"playerctl",
		"bluez",
		"bluez-utils",
		"alacritty",
		"wl-clipboard",

		"vlc",
		"rhythmbox",

		"polkit",
		"sway",
		"swaybg",
		"j4-dmenu-desktop",
		"bemenu",
		"grim",
		"wf-recorder",
		"slurp",
		"swaylock",
		"xdg-desktop-portal-wlr",

		"i3status",
		"dmenu",

		"openssh",
		"btop",
		"grub-btrfs",

		"bitwarden-cli",
		//"brscan4",
		//"brother-dcp1510",
		// "leiningen" -- for dactyl development
		// "clojure" -- for dactyl development

	}
}

func (y Yay) CustomPackageList(pkgs []string) api.Package {
	return yayPackage(pkgs)
}

func (y Yay) EnsurePackagerAction(homedir string) action.Object {
	return action.WithCondition{
		If: action.Not(action.PathExists(path.Join(homedir, "yay"))),
		Then: action.List{
			action.ShellCommand("git", "clone", "https://aur.archlinux.org/yay.git", path.Join(homedir, "/yay")),
			action.ShellCommand("bash", "-c", "cd ~/yay && makepkg -si"),
		},
	}
}
