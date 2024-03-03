package macbook

import (
	"os"
	"path"

	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/common"
)

var (
	homeDir    = os.Getenv("HOME")
	expoConfig = common.ExpoConfig
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
		expoConfig.ExpoSdkGl(path.Join(homeDir, "/expo/expo/packages/expo-gl")),
		common.HomeWorkspace,
		common.DotfilesWorkspace,
	},
	Init: []env.InitAction{
		{Args: []string{"mycli", "api", "--simple", "backup:zsh_history"}},
	},
	CustomSetupAction: func(ctx env.Context) error {
		return nil
	},
}
