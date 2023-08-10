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

func PackageInstallAction(packages []Package) action.Object {
	packageNames := []string{}
	for _, pkg := range packages {
		packageNames = append(packageNames, fmt.Sprintf(" - %s", pkg.String()))
	}
	label := fmt.Sprintf("Install system packages:\n%s", strings.Join(packageNames, "\n"))
	return action.SimpleAction{
		Run: func() error {
			for _, pkg := range packages {
				if err := pkg.Install(); err != nil {
					return err
				}
			}
			return nil
		},
		Label: label,
	}
}
