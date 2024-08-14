package work

import (
	"os"
	"path"

	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/common"
)

var (
	homeDir            = os.Getenv("HOME")
	expoConfig         = common.ExpoConfig
	expoLauncherConfig = common.ExpoLauncherConfig(path.Join(homeDir, "expo"))
)

var Config = env.EnvironmentConfig{
	Workspaces: []env.Workspace{
		expoConfig.LegacyExpoCli(path.Join(homeDir, "expo/expo-cli")),
		expoConfig.EasCli(path.Join(homeDir, "expo/eas-cli")),
		expoConfig.EasBuild(path.Join(homeDir, "expo/eas-build")),
		expoConfig.Turtle(path.Join(homeDir, "expo/turtle-v2")),
		expoConfig.UniverseWWW(path.Join(homeDir, "expo/universe/server/www")),
		expoConfig.UniverseWebsite(path.Join(homeDir, "expo/universe/server/website")),
		expoConfig.TurtleClassic(path.Join(homeDir, "expo/turtle")),
		expoConfig.ExpoSdk(path.Join(homeDir, "expo/expo")),
		expoConfig.ExpoSdkGl(path.Join(homeDir, "expo/expo/packages/expo-gl")),
		expoConfig.EASBuildCache(path.Join(homeDir, "expo/eas-build-cache")),
		common.DotfilesWorkspace,
		common.HomeWorkspace,
		common.MembraneConfig.VideoCompositor(path.Join(homeDir, "membrane/live_compositor")),
		common.MembraneConfig.VideoCompositorTypescript(path.Join(homeDir, "membrane/live_compositor/ts")),
	},
	Actions: []env.LauncherAction{
		expoLauncherConfig.EasCli,
		expoLauncherConfig.ExpoCliRebuild,
		expoLauncherConfig.ExpoDocs,
		expoLauncherConfig.Submit,
		expoLauncherConfig.Turtle,
		expoLauncherConfig.Submit,
		expoLauncherConfig.UniverseWWW,
		expoLauncherConfig.UniverseWWWUnit,
		expoLauncherConfig.UniverseWebsite,
		expoLauncherConfig.UniverseWebsiteInternal,
	},
	Init: []env.InitAction{
		{Args: []string{"google-chrome-stable", "--proxy-pac-url=http://localhost:2000/proxy.pac"}},
		{Args: []string{"slack"}},
		{Args: []string{"alacritty", "--class", "workspace2"}},
		{Args: []string{"alacritty", "--class", "workspace6"}},
		{Args: []string{"mycli", "api", "--simple", "backup:zsh_history"}},
	},
	DockerEnvsSpec: []env.DockerEnvSpec{
		{
			Name:           "compositor",
			ImageName:      "mycli-compositor",
			DockerfilePath: path.Join(homeDir, ".dotfiles/configs/dockerfiles/compositor.Dockerfile"),
			ContainerName:  "live-compositor",
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
