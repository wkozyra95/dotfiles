package common

import (
	"path"

	"github.com/wkozyra95/dotfiles/env"
)

type MembraneWorkspacesConfig struct {
	VideoCompositor           func(p string) env.Workspace
	VideoCompositorTypescript func(p string) env.Workspace
}

var MembraneConfig = MembraneWorkspacesConfig{
	VideoCompositor: func(p string) env.Workspace {
		return env.Workspace{
			Name: "live_compositor",
			Path: p,
			VimConfig: env.VimConfig{
				FiletypeConfig: map[string]env.VimFiletypeConfig{
					"json": {
						IndentSize: 4,
					},
				},
				JsonlsSchemas: []env.JsonSchema{
					{
						FileMatch: []string{"*.scene.json"},
						Url:       "file://" + path.Join(p, "schemas/scene.schema.json"),
					},
					{
						FileMatch: []string{"*.register.json"},
						Url:       "file://" + path.Join(p, "schemas/register.schema.json"),
					},
				},
			},
		}
	},
	VideoCompositorTypescript: func(p string) env.Workspace {
		return env.Workspace{
			Name: "live_compositor",
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
