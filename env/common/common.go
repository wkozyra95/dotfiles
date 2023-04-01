package common

import (
	"os"
	"path"

	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/logger"
)

var homeDir = os.Getenv("HOME")

var log = logger.NamedLogger("common")

var DotfilesWorkspace = env.Workspace{
	Name: "dotfiles",
	Path: path.Join(homeDir, "/.dotfiles"),
	VimConfig: env.VimConfig{
		GoEfm: map[string]interface{}{
			"formatCommand": "golines --max-len=120 --base-formatter=\"gofumpt\"",
			"formatStdin":   true,
		},
		Actions: []env.VimAction{
			{
				Id:   "dotfiles_go_build",
				Name: "[workspace] build",
				Args: []string{"make"},
				Cwd:  path.Join(homeDir, ".dotfiles"),
			},
		},
	},
}

var HomeWorkspace = env.Workspace{Name: "home", Path: homeDir}
