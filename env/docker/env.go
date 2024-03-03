package docker

import (
	"os"

	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/platform/ubuntu"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/common"
)

var homeDir = os.Getenv("HOME")

var Config = env.EnvironmentConfig{
	Workspaces: []env.Workspace{
		common.DotfilesWorkspace,
	},
	Actions: []env.LauncherAction{},
	Init:    []env.InitAction{},
	CustomSetupAction: func(ctx env.Context) error {
		pkgInstaller := ubuntu.Apt{}
		if err := pkgInstaller.EnsurePackagerInstalled(homeDir); err != nil {
			return err
		}
		pkgs := []api.Package{pkgInstaller.CustomPackageList([]string{})}
		if err := api.InstallPackages(pkgs); err != nil {
			return err
		}
		return nil
	},
}
