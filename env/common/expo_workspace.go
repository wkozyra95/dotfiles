package common

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/wkozyra95/dotfiles/env"
)

var (
	eslintConfigEnabled = true
	EslintConfig        = env.VimConfig{Eslint: &eslintConfigEnabled}
	turtleDatabaseUrls  = map[string]string{
		"dev":        "LOCAL_DATABASE_URL",
		"staging":    "STAGING_DATABASE_URL",
		"production": "PRODUCTION_DATABASE_URL",
	}
)

func YarnBuild(p string) env.VimAction {
	return env.VimAction{
		Id:   "yarn build",
		Name: "[workspace] yarn build",
		Args: []string{"bash", "-c", "yarn && yarn build"},
		Cwd:  p,
	}
}

func YarnLint(p string) env.VimAction {
	return env.VimAction{
		Id:   "yarn lint",
		Name: "[workspace] yarn lint",
		Args: []string{"bash", "-c", "yarn && yarn lint"},
		Cwd:  p,
	}
}

type ExpoWorkspacesConfig struct {
	LegacyExpoCli   func(p string) env.Workspace
	EasCli          func(p string) env.Workspace
	EasBuild        func(p string) env.Workspace
	Turtle          func(p string) env.Workspace
	UniverseWWW     func(p string) env.Workspace
	UniverseWebsite func(p string) env.Workspace
	TurtleClassic   func(p string) env.Workspace
	ExpoSdk         func(p string) env.Workspace
	ExpoSdkGl       func(p string) env.Workspace
	EASBuildCache   func(p string) env.Workspace
}

func tryReadingDatabaseSecrets(file string) map[string]string {
	dbMap := map[string]string{}
	content, contentErr := os.ReadFile(file)
	if contentErr != nil {
		log.Error(contentErr.Error())
	} else {
		for key, entry := range turtleDatabaseUrls {
			rg := regexp.MustCompile(fmt.Sprintf("%s=(.*)", regexp.QuoteMeta(entry)))
			matches := rg.FindSubmatchIndex(content)
			if matches != nil {
				dbMap[key] = string(content[matches[2]:matches[3]])
			}
		}
	}
	return dbMap
}

var ExpoConfig = ExpoWorkspacesConfig{
	LegacyExpoCli: func(p string) env.Workspace {
		return env.Workspace{
			Name: "expo-cli", Path: p, VimConfig: env.VimConfig{
				Eslint: EslintConfig.Eslint,
				Actions: []env.VimAction{
					YarnBuild(p),
					YarnLint(p),
				},
			},
		}
	},
	EasCli: func(p string) env.Workspace {
		return env.Workspace{Name: "eas-cli", Path: p, VimConfig: env.VimConfig{
			Eslint: EslintConfig.Eslint,
			Actions: []env.VimAction{
				YarnBuild(p),
				YarnLint(p),
			},
		}}
	},
	EasBuild: func(p string) env.Workspace {
		return env.Workspace{Name: "eas-build", Path: p, VimConfig: env.VimConfig{
			Eslint: EslintConfig.Eslint,
			Actions: []env.VimAction{
				YarnBuild(p),
				YarnLint(p),
			},
		}}
	},
	Turtle: func(p string) env.Workspace {
		return env.Workspace{
			Name: "turtle", Path: p, VimConfig: env.VimConfig{
				Eslint: EslintConfig.Eslint,
				Actions: []env.VimAction{
					YarnBuild(p),
					YarnLint(p),
				},
				Databases: env.LazyValue[map[string]string](
					func() map[string]string { return tryReadingDatabaseSecrets(path.Join(p, "database/secrets.env")) },
				),
			},
		}
	},
	UniverseWWW: func(p string) env.Workspace {
		return env.Workspace{Name: "www", Path: p, VimConfig: env.VimConfig{
			Eslint: EslintConfig.Eslint,
			Databases: env.LazyValue[map[string]string](
				func() map[string]string { return tryReadingDatabaseSecrets(path.Join(homeDir, ".secrets/www_db.env")) },
			),
		}}
	},
	UniverseWebsite: func(p string) env.Workspace {
		return env.Workspace{Name: "website", Path: p, VimConfig: EslintConfig}
	},
	TurtleClassic: func(p string) env.Workspace {
		return env.Workspace{Name: "classic", Path: p, VimConfig: EslintConfig}
	},
	ExpoSdk: func(p string) env.Workspace {
		return env.Workspace{Name: "sdk", Path: p, VimConfig: EslintConfig}
	},
	ExpoSdkGl: func(p string) env.Workspace {
		return env.Workspace{Name: "gl-cpp", Path: p, VimConfig: env.VimConfig{
			Eslint: EslintConfig.Eslint,
			CmakeEfm: map[string]interface{}{
				"formatCommand": "cmake-format --tab-size 4 ${INPUT}",
				"formatStdin":   false,
			},
			Actions: []env.VimAction{
				{
					Id:   "expo-gl-build-cpp",
					Name: "[workspace] build cpp",
					Args: []string{"./gradlew", ":expo-gl:buildCMakeDebug"},
					Cwd:  path.Join(p, "../../android"),
				},
			},
		}}
	},
	EASBuildCache: func(p string) env.Workspace {
		return env.Workspace{
			Name: "cache",
			Path: p,
			VimConfig: env.VimConfig{
				GoEfm: map[string]interface{}{
					"formatCommand": "gofumpt",
					"formatStdin":   true,
				},
			},
		}
	},
}
