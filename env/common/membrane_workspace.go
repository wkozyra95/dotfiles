package common

import (
	"github.com/wkozyra95/dotfiles/env"
)

type MembraneWorkspacesConfig struct {
	VideoCompositor func(p string) env.Workspace
}

var MembraneConfig = MembraneWorkspacesConfig{
	VideoCompositor: func(p string) env.Workspace {
		return env.Workspace{
			Name: "video_compositor",
			Path: p,
			VimConfig: env.VimConfig{
				FiletypeConfig: map[string]any{
					"json": env.VimFiletypeConfig{
						IndentSize: 4,
					},
				},
			},
		}
	},
}
