package docker

import (
	"os"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/common"
)

var homeDir = os.Getenv("HOME")

var Config = env.EnvironmentConfig{
	Workspaces: []env.Workspace{
		common.DotfilesWorkspace,
	},
	Actions: []env.LauncherAction{},
	Init:    []env.InitAction{},
	CustomSetupAction: func(ctx env.Context) action.Object {
		return action.Nop()
	},
}
