package common

import (
	"path"

	"github.com/wkozyra95/dotfiles/env"
)

type SmelterLauncherConfigType struct {
	Smelter env.LauncherAction
}

func SmelterLauncherConfig(p string) SmelterLauncherConfigType {
	shell := func(id string, dir string, workspaceID int) env.LauncherTask {
		return env.LauncherTask{
			Id:           id,
			Cwd:          path.Join(p, dir),
			Args:         []string{"zsh"},
			RunAsService: true,
			WorkspaceID:  workspaceID,
		}
	}
	pnpmDev := func(id string, dir string, workspaceID int) env.LauncherTask {
		return env.LauncherTask{
			Id:           id,
			Cwd:          path.Join(p, dir),
			Args:         []string{"pnpm", "dev"},
			RunAsService: true,
			WorkspaceID:  workspaceID,
		}
	}

	return SmelterLauncherConfigType{
		Smelter: env.LauncherAction{
			Id: "smelter",
			Tasks: []env.LauncherTask{
				shell("smelter-shell-ws2", "smelter", env.Workspace2),
				shell("smelter-2-shell-ws3", "smelter-2", env.Workspace3),
				shell("smelter-website-shell-ws4", "smelter-website", env.Workspace4),
				pnpmDev("smelter-website-dev-ws4", "smelter-website", env.Workspace4),
				shell("smelter-tools-shell-ws5", "tools", env.Workspace5),
				pnpmDev("smelter-tools-dev-ws5", "tools", env.Workspace5),
				shell("smelter-shell-ws6", "smelter", env.Workspace6),
				shell("smelter-2-shell-ws7", "smelter-2", env.Workspace7),
			},
		},
	}
}
