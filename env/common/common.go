package common

import (
	"os"
	"path"

	"github.com/wkozyra95/dotfiles/env"
)

var homeDir = os.Getenv("HOME")

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
				ID:   "dotfiles_go_build",
				Name: "[workspace] build",
				Args: []string{"make"},
				Cwd:  path.Join(homeDir, ".dotfiles"),
			},
		},
	},
}

var HomeWorkspace = env.Workspace{Name: "home", Path: homeDir}

// DotfilesSessionTemplate opens a single shell in the dotfiles directory
// (standalone unit — no partner workspace).
var DotfilesSessionTemplate = env.SessionTemplate{
	ID:          "dotfiles",
	Name:        "dotfiles",
	DefaultName: "dotfiles",
	Tasks: []env.SessionTask{
		{ID: "shell", Cwd: path.Join(homeDir, ".dotfiles"), Args: []string{"zsh"}, Slot: 0},
	},
}

// SkillsSessionTemplate opens a single shell in the ~/skills repo.
var SkillsSessionTemplate = env.SessionTemplate{
	ID:   "skills",
	Name: "skills",
	Tasks: []env.SessionTask{
		{ID: "shell", Cwd: path.Join(homeDir, "skills"), Args: []string{"zsh"}, Slot: 0},
	},
}
