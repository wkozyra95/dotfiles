package work

import (
	"os"
	"path"

	"github.com/wkozyra95/dotfiles/api/sway"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/common"
)

var homeDir = os.Getenv("HOME")

var Config = env.EnvironmentConfig{
	Workspaces: []env.Workspace{
		common.DotfilesWorkspace,
		common.HomeWorkspace,
		common.SmelterConfig.Smelter(path.Join(homeDir, "smelter/smelter")),
		common.SmelterConfig.SmelterTypescript(path.Join(homeDir, "smelter/smelter/ts")),
		common.SmelterConfig.Smelter(path.Join(homeDir, "smelter/smelter-2")),
		common.SmelterConfig.SmelterTypescript(path.Join(homeDir, "smelter/smelter-2/ts")),
	},
	SessionTemplates: func() []env.SessionTemplate {
		templates := []env.SessionTemplate{common.DotfilesSessionTemplate}
		templates = append(templates, common.SmelterSessionTemplates(path.Join(homeDir, "smelter"))...)
		return append(templates, common.SkillsSessionTemplate)
	},
	Init: []env.InitAction{
		{Args: []string{"google-chrome-stable", "--proxy-pac-url=http://localhost:2000/proxy.pac"}},
		{Args: []string{"slack"}},
		{Args: []string{"mycli", "api", "--simple", "backup:zsh_history"}},
	},
	SwayHandlers: []sway.Handler{
		sway.WorkspaceSyncHandler([]sway.WorkspacePair{
			{
				A: sway.WorkspaceBinding{Workspace: "2", Output: "DP-2"},
				B: sway.WorkspaceBinding{Workspace: "6", Output: "DP-3"},
			},
			{
				A: sway.WorkspaceBinding{Workspace: "3", Output: "DP-2"},
				B: sway.WorkspaceBinding{Workspace: "7", Output: "DP-3"},
			},
		}),
		sway.MoveAppToVisibleWorkspaceHandler("ffplay", []string{"6", "7"}),
		sway.PortraitSplitvHandler(),
	},
	DockerEnvsSpec: []env.DockerEnvSpec{
		{
			Name:           "smelter",
			ImageName:      "mycli-smelter",
			DockerfilePath: path.Join(homeDir, ".dotfiles/configs/dockerfiles/smelter.Dockerfile"),
			ContainerName:  "smelter",
		},
	},
	Backup: env.BackupConfig{
		GpgKeyring: true,
		Secrets: map[string]string{
			path.Join(homeDir, ".secrets"): "secrets",
			path.Join(homeDir, ".ssh"):     "ssh",
		},
		Data: map[string]string{
			path.Join(homeDir, ".secrets"): "secrets",
			path.Join(homeDir, ".ssh"):     "ssh",
		},
	},
	CustomSetupAction: func(ctx env.Context) error {
		return nil
	},
}
