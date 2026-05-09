package common

import (
	"path"

	"github.com/wkozyra95/dotfiles/env"
)

type SmelterWorkspacesConfig struct {
	Smelter           func(p string) env.Workspace
	SmelterTypescript func(p string) env.Workspace
}

var SmelterConfig = SmelterWorkspacesConfig{
	Smelter: func(p string) env.Workspace {
		return env.Workspace{
			Name: "smelter",
			Path: p,
			VimConfig: env.VimConfig{
				FiletypeConfig: map[string]env.VimFiletypeConfig{
					"json": {
						IndentSize: 4,
					},
				},
				JsonlsSchemas: []env.JSONSchema{
					{
						FileMatch: []string{"*.scene.json"},
						URL:       "file://" + path.Join(p, "schemas/scene.schema.json"),
					},
					{
						FileMatch: []string{"*.register.json"},
						URL:       "file://" + path.Join(p, "schemas/register.schema.json"),
					},
				},
			},
		}
	},
	SmelterTypescript: func(p string) env.Workspace {
		return env.Workspace{
			Name: "smelter",
			Path: p,
			VimConfig: env.VimConfig{
				FiletypeConfig: map[string]env.VimFiletypeConfig{
					"json": {
						IndentSize: 2,
					},
				},
			},
		}
	},
}
