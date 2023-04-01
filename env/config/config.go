package config

import (
	"os"

	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/docker"
	"github.com/wkozyra95/dotfiles/env/home"
	"github.com/wkozyra95/dotfiles/env/homeoffice"
	"github.com/wkozyra95/dotfiles/env/macbook"
	"github.com/wkozyra95/dotfiles/env/work"
	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("api")

func GetConfig() env.EnvironmentConfig {
	switch os.Getenv("CURRENT_ENV") {
	case "home":
		return home.Config
	case "work":
		return work.Config
	case "homeoffice":
		return homeoffice.Config
	case "macbook":
		return macbook.Config
	case "docker":
		return docker.Config
	default:
		log.Warn("Missing or invalid CURRENT_ENV")
		return env.EnvironmentConfig{
			Workspaces: []env.Workspace{},
			Actions:    []env.LauncherAction{},
			Init:       []env.InitAction{},
		}
	}
}
