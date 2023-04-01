package api

import (
	"fmt"
	"strings"

	"github.com/wkozyra95/dotfiles/action"
)

type Package interface {
	Install() error
	String() string
}

type PackageInstaller interface {
	EnsurePackagerAction(homedir string) action.Object
	UpgradePackages() error
	DevelopmentTools() Package
	ShellTools() Package
	Desktop() Package
}

var PackageInstallAction = action.SimpleActionBuilder[[]Package]{
	CreateRun: func(pkgs []Package) func() error {
		return func() error {
			for _, pkg := range pkgs {
				if err := pkg.Install(); err != nil {
					return err
				}
			}
			return nil
		}
	},
	String: func(pkgs []Package) string {
		packages := []string{}
		for _, pkg := range pkgs {
			packages = append(packages, fmt.Sprintf(" - %s", pkg.String()))
		}
		return fmt.Sprintf("Install system packages:\n%s", strings.Join(packages, "\n"))
	},
}.Init()
