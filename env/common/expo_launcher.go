package common

import (
	"path"

	"github.com/wkozyra95/dotfiles/env"
)

type ExpoLauncherConfigType struct {
	UniverseWWW             env.LauncherAction
	UniverseWWWUnit         env.LauncherAction
	UniverseWebsite         env.LauncherAction
	UniverseWebsiteInternal env.LauncherAction
	EasCli                  env.LauncherAction
	Turtle                  env.LauncherAction
	Submit                  env.LauncherAction
	ExpoCliRebuild          env.LauncherAction
	ExpoDocs                env.LauncherAction
	ExpoGL                  env.LauncherAction
}

func ExpoLauncherConfig(p string) ExpoLauncherConfigType {
	universeWWWDockerUp := env.LauncherTask{
		Id:   "www-docker-up",
		Cwd:  path.Join(p, "universe/server/www"),
		Args: []string{"yarn", "docker-up"},
	}
	universeWWWYarn := env.LauncherTask{
		Id:   "www-docker-yarn",
		Cwd:  path.Join(p, "universe/server/www"),
		Args: []string{"yarn"},
	}
	universeWWWStart := env.LauncherTask{
		Id:           "www-docker-start",
		Cwd:          path.Join(p, "universe/server/www"),
		Args:         []string{"yarn", "start:docker"},
		RunAsService: true,
		WorkspaceID:  env.Workspace6,
	}
	easBuildLibsWatch := env.LauncherTask{
		Id:           "eas-build-watch",
		Cwd:          path.Join(p, "eas-build"),
		Args:         []string{"yarn", "watch"},
		RunAsService: true,
		WorkspaceID:  env.Workspace6,
	}
	turtleDockerUp := env.LauncherTask{
		Id:   "turtle-docker-up",
		Cwd:  path.Join(p, "turtle-v2"),
		Args: []string{"yarn", "docker:up"},
	}

	return ExpoLauncherConfigType{
		UniverseWWW: env.LauncherAction{
			Id: "www",
			Tasks: []env.LauncherTask{
				universeWWWYarn,
				universeWWWDockerUp,
				universeWWWStart,
			},
		},
		UniverseWWWUnit: env.LauncherAction{
			Id: "www-unit",
			Tasks: []env.LauncherTask{
				{
					Id:   "www-unit",
					Cwd:  path.Join(p, "universe/server/www"),
					Args: []string{"yarn", "jest-unit"},
				},
			},
		},
		UniverseWebsite: env.LauncherAction{
			Id: "website",
			Tasks: []env.LauncherTask{
				universeWWWYarn,
				universeWWWDockerUp,
				universeWWWStart,
				{
					Id:   "website-yarn",
					Cwd:  path.Join(p, "universe/server/website"),
					Args: []string{"yarn"},
				},
				{
					Id:           "website-start",
					Cwd:          path.Join(p, "universe/server/website"),
					Args:         []string{"direnv", "exec", ".", "yarn", "start:local"},
					RunAsService: true,
					WorkspaceID:  env.Workspace6,
				},
			},
		},
		UniverseWebsiteInternal: env.LauncherAction{
			Id: "website",
			Tasks: []env.LauncherTask{
				universeWWWYarn,
				universeWWWDockerUp,
				{
					Id:   "website-yarn",
					Cwd:  path.Join(p, "universe/server/internal"),
					Args: []string{"yarn"},
				},
				universeWWWStart,
				{
					Id:          "website-start",
					Cwd:         path.Join(p, "universe/server/internal"),
					Args:        []string{"yarn", "dev"},
					WorkspaceID: env.Workspace6,
				},
			},
		},
		EasCli: env.LauncherAction{
			Id: "cli",
			Tasks: []env.LauncherTask{
				{
					Id:           "eas-cli-watch",
					Cwd:          path.Join(p, "eas-cli"),
					Args:         []string{"yarn", "watch"},
					RunAsService: true,
					WorkspaceID:  env.Workspace6,
				},
				easBuildLibsWatch,
			},
		},
		Turtle: env.LauncherAction{
			Id: "turtle",
			Tasks: []env.LauncherTask{
				turtleDockerUp,
				easBuildLibsWatch,
				{
					Id:           "turtle-libs-watch",
					Cwd:          path.Join(p, "turtle-v2"),
					Args:         []string{"yarn", "watch:libs"},
					RunAsService: true,
					WorkspaceID:  env.Workspace7,
				},
				{
					Id:           "turtle-start-api",
					Cwd:          path.Join(p, "turtle-v2/src/services/turtle-api"),
					Args:         []string{"yarn", "start"},
					RunAsService: true,
					WorkspaceID:  env.Workspace7,
				},
				{
					Id:           "turtle-start-scheduler",
					Cwd:          path.Join(p, "turtle-v2/src/services/scheduler"),
					Args:         []string{"yarn", "start"},
					RunAsService: true,
					WorkspaceID:  env.Workspace7,
				},
				{
					Id:           "turtle-start-launcher",
					Cwd:          path.Join(p, "turtle-v2/src/services/launcher"),
					Args:         []string{"yarn", "start"},
					RunAsService: true,
					WorkspaceID:  env.Workspace7,
				},
				{
					Id:           "turtle-start-synchronizer",
					Cwd:          path.Join(p, "turtle-v2/src/services/synchronizer"),
					Args:         []string{"yarn", "start"},
					RunAsService: true,
					WorkspaceID:  env.Workspace7,
				},
			},
		},
		Submit: env.LauncherAction{
			Id: "submit",
			Tasks: []env.LauncherTask{
				easBuildLibsWatch,
				turtleDockerUp,
				{
					Id:           "turtle-libs-watch",
					Cwd:          path.Join(p, "turtle-v2"),
					Args:         []string{"yarn", "watch:libs"},
					RunAsService: true,
					WorkspaceID:  env.Workspace7,
				},
				{
					Id:           "turtle-start-submit",
					Cwd:          path.Join(p, "turtle-v2/src/services/submission-service"),
					Args:         []string{"yarn", "start"},
					RunAsService: true,
					WorkspaceID:  env.Workspace7,
				},
			},
		},
		ExpoCliRebuild: env.LauncherAction{
			Id: "expo-cli-rebuild",
			Tasks: []env.LauncherTask{
				{
					Id:   "expo-cli-rebuild-config-types",
					Cwd:  path.Join(p, "expo/packages/@expo/config-types"),
					Args: []string{"yarn", "build"},
				},
				{
					Id:   "expo-cli-rebuild-config-plugins",
					Cwd:  path.Join(p, "expo/packages/@expo/config-plugins"),
					Args: []string{"yarn", "build"},
				},
				{
					Id:   "expo-cli-rebuild-config",
					Cwd:  path.Join(p, "expo/packages/@expo/config"),
					Args: []string{"yarn", "build"},
				},
			},
		},
		ExpoDocs: env.LauncherAction{
			Id: "docs",
			Tasks: []env.LauncherTask{
				{
					Id:           "expo-docs",
					Cwd:          path.Join(p, "expo/docs"),
					Args:         []string{"yarn", "dev"},
					RunAsService: true,
				},
			},
		},
		ExpoGL: env.LauncherAction{
			Id: "gl",
			Tasks: []env.LauncherTask{
				{
					Id:   "expo-gl-js-build",
					Cwd:  path.Join(p, "expo/packages/expo-gl"),
					Args: []string{"yarn", "build"},
				},
				{
					Id:   "expo-gl-cpp",
					Cwd:  path.Join(p, "expo/android"),
					Args: []string{"gradlew", ":expo-gl:buildCMakeDebug"},
				},
			},
		},
	}
}
