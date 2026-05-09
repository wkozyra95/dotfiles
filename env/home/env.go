package home

import (
	"os"
	"path"

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
					{
						ID:   "cargo_watch",
						Name: "[workspace] cargo watch (new terminal)",
						Args: []string{"mycli", "launch", "--job", "npm-cache"},
						Cwd:  path.Join(homeDir, "drive/MyProjects/npm-cache"),
					},
					{
						ID:   "cargo_watch_run",
						Name: "[workspace] cargo run (new terminal)",
						Args: []string{"mycli", "launch", "--job", "npm-cache-run"},
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
		{Name: "test", Path: path.Join(homeDir, "playground/vimtest"), VimConfig: env.VimConfig{
			Eslint: common.EslintConfig.Eslint,
			CmakeEfm: map[string]interface{}{
				"formatCommand": "cmake-format --tab-size 4 ${INPUT}",
				"formatStdin":   false,
			},
		}},
		{
			Name: "cache",
			Path: path.Join(homeDir, "drive/MyProjects/eas-build-cache"),
			VimConfig: env.VimConfig{
				GoEfm: map[string]interface{}{
					"formatCommand": "gofumpt",
					"formatStdin":   true,
				},
			},
		},
		{
			Name: "expo-rust",
			Path: path.Join(homeDir, "drive/MyProjects/expo-rust"),
			VimConfig: env.VimConfig{
				Eslint: common.EslintConfig.Eslint,
			},
		},
		{
			Name: "expo-myapp",
			Path: path.Join(homeDir, "drive/MyProjects/myapp"),
			VimConfig: env.VimConfig{
				Eslint: common.EslintConfig.Eslint,
			},
		},
		common.MembraneConfig.VideoCompositor(path.Join(homeDir, "playground/video_compositor")),
	},
	Actions: []env.LauncherAction{
		{
			ID: "debug",
			Tasks: []env.LauncherTask{
				{
					ID:           "debug",
					Cwd:          path.Join(homeDir, "playground"),
					Args:         []string{"zsh", "-c", "sleep 10 && exit 1"},
					RunAsService: true,
					WorkspaceID:  env.Workspace3,
				},
				{
					ID:           "debug1",
					Cwd:          path.Join(homeDir, "playground"),
					Args:         []string{"zsh", "-c", "lskadjfsld;j"},
					RunAsService: true,
					WorkspaceID:  env.Workspace4,
				},
				{
					ID:           "debug2",
					Cwd:          path.Join(homeDir, "playground"),
					Args:         []string{"htop"},
					RunAsService: true,
					WorkspaceID:  env.Workspace5,
				},
			},
		},
		{
			ID: "npm-cache-run",
			Tasks: []env.LauncherTask{
				{
					ID:           "npm-watch-run-cargo",
					Args:         []string{"cargo", "watch", "-x", "run"},
					Cwd:          path.Join(homeDir, "drive/MyProjects/npm-cache"),
					RunAsService: true,
				},
			},
		},
		{
			ID: "npm-cache",
			Tasks: []env.LauncherTask{
				{
					ID:           "npm-watch-cargo",
					Args:         []string{"cargo", "watch"},
					Cwd:          path.Join(homeDir, "drive/MyProjects/npm-cache"),
					RunAsService: true,
				},
			},
		},
	},
	Init: []env.InitAction{
		{Args: []string{"alacritty", "--class", "workspace6"}},
		{Args: []string{"firefox"}},
		{Args: []string{"mycli", "api", "--simple", "backup:zsh_history"}},
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
			Name:           "compositor",
			ImageName:      "mycli-compositor",
			DockerfilePath: path.Join(homeDir, ".dotfiles/configs/dockerfiles/compositor.Dockerfile"),
			ContainerName:  "live-compositor",
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
	Actions:        []env.LauncherAction{},
	Init:           []env.InitAction{},
	Backup:         env.BackupConfig{},
	DockerEnvsSpec: []env.DockerEnvSpec{},
	CustomSetupAction: func(ctx env.Context) error {
		return nil
	},
}
