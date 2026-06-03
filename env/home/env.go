package home

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
		{
			Name: "npm-cache", Path: path.Join(homeDir, "drive/MyProjects/npm-cache"),
			VimConfig: env.VimConfig{
				Actions: []env.VimAction{
					{
						ID:   "cargo_build",
						Name: "[workspace] cargo build",
						Args: []string{"cargo", "build"},
						Cwd:  path.Join(homeDir, "drive/MyProjects/npm-cache"),
					},
				},
			},
		},
		{
			Name: "dactyl-model", Path: path.Join(homeDir, "/drive/MyProjects/dactyl/dactyl-keyboard"),
			VimConfig: env.VimConfig{
				Actions: []env.VimAction{
					{
						ID:   "dactyl_build",
						Name: "[workspace] build",
						Args: []string{"lein", "run", "src/dactyl_keyboard/dactyl.clj"},
						Cwd:  path.Join(homeDir, "/drive/MyProjects/dactyl/dactyl-keyboard"),
					},
				},
			},
		},
		common.HomeWorkspace,
		common.SmelterConfig.Smelter(path.Join(homeDir, "smelter/smelter")),
		common.SmelterConfig.SmelterTypescript(path.Join(homeDir, "smelter/smelter/ts")),
	},
	SessionTemplates: func() []env.SessionTemplate {
		return append(
			common.SmelterSessionTemplates(path.Join(homeDir, "smelter")),
			common.DotfilesSessionTemplate,
			common.SkillsSessionTemplate,
		)
	},
	Init: []env.InitAction{
		{Args: []string{"firefox"}},
		{Args: []string{"mycli", "api", "--simple", "backup:zsh_history"}},
	},
	SwayHandlers: []sway.Handler{
		sway.MoveAppToVisibleWorkspaceHandler("ffplay", []string{"6", "7"}),
		sway.PortraitSplitvHandler(),
	},
	Backup: env.BackupConfig{
		GpgKeyring: true,
		Secrets: map[string]string{
			path.Join(homeDir, ".secrets"): "secrets",
			path.Join(homeDir, ".ssh"):     "ssh",
		},
		Data: map[string]string{
			path.Join(homeDir, ".secrets"):     "secrets",
			path.Join(homeDir, ".ssh"):         "ssh",
			path.Join(homeDir, ".zsh_history"): "zsh_history",
			path.Join(homeDir, "drive"):        "drive",
		},
	},
	DockerEnvsSpec: []env.DockerEnvSpec{
		{
			Name:           "ubuntu",
			ImageName:      "mycli-ubuntu-image",
			DockerfilePath: path.Join(homeDir, ".dotfiles/configs/dockerfiles/ubuntu.Dockerfile"),
			ContainerName:  "ubuntu",
		},
		{
			Name:           "expo-sdk",
			ImageName:      "mycli-expo-sdk-image",
			DockerfilePath: path.Join(homeDir, ".dotfiles/configs/dockerfiles/expo-sdk.Dockerfile"),
			ContainerName:  "expo-sdk",
		},
		{
			Name:           "smelter",
			ImageName:      "mycli-smelter",
			DockerfilePath: path.Join(homeDir, ".dotfiles/configs/dockerfiles/smelter.Dockerfile"),
			ContainerName:  "smelter",
		},
	},
	CustomSetupAction: func(ctx env.Context) error {
		return nil
	},
}

var NasConfig = env.EnvironmentConfig{
	Workspaces: []env.Workspace{
		common.DotfilesWorkspace,
	},
	Init:           []env.InitAction{},
	Backup:         env.BackupConfig{},
	DockerEnvsSpec: []env.DockerEnvSpec{},
	CustomSetupAction: func(ctx env.Context) error {
		return nil
	},
}
