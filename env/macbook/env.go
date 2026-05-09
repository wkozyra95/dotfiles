package macbook

import (
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/common"
)

var Config = env.EnvironmentConfig{
	Workspaces: []env.Workspace{
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
