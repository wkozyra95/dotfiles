package macos

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

var log = logger.NamedLogger("brew")

var _ api.PackageInstaller = Brew{}

type brewPackage []string

func (p brewPackage) Install() error {
	for _, pkg := range p {
		var stdout bytes.Buffer
		checkErr := exec.
			Command().
			WithBufout(&stdout, &bytes.Buffer{}).
			WithEnv("HOMEBREW_NO_AUTO_UPDATE=1").
			Args("brew", "list", pkg).Run()
		if err := checkErr; err == nil {
			continue
		}
		installErr := exec.
			Command().
			WithStdio().
			WithEnv("HOMEBREW_NO_AUTO_UPDATE=1").
			Args("brew", "install", pkg).Run()
		if err := installErr; err != nil {
			return err
		}
	}
	return nil
}

func (p brewPackage) String() string {
	return fmt.Sprintf("brew install %s", strings.Join(p, " "))
}

type Brew struct{}

func (y Brew) UpgradePackages() error {
	log.Warn("not implemented")
	return nil
}

func (y Brew) DevelopmentTools() api.Package {
	return brewPackage{"cmake", "go", "ninja"}
}

func (y Brew) ShellTools() api.Package {
	return brewPackage{"neovim", "htop", "ripgrep", "the_silver_searcher", "fzf", "git", "git-crypt"}
}

func (y Brew) Desktop() api.Package {
	return brewPackage{}
}

func (y Brew) EnsurePackagerAction(homedir string) action.Object {
	return action.WithCondition{
		If: action.Not(action.CommandExists("brew")),
		Then: action.List{
			action.ShellCommand("curl", "-fsSL", "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"),
		},
	}
}
