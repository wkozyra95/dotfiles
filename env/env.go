package env

import (
	"encoding/json"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("common")

const (
	Workspace1  int = 1
	Workspace2      = 2
	Workspace3      = 3
	Workspace4      = 4
	Workspace5      = 5
	Workspace6      = 6
	Workspace7      = 7
	Workspace8      = 8
	Workspace9      = 9
	Workspace10     = 10
)

type VimConfig struct {
	GoEfm     map[string]interface{}       `json:"go_efm,omitempty"`
	CmakeEfm  map[string]interface{}       `json:"cmake_efm,omitempty"`
	Eslint    *bool                        `json:"eslint,omitempty"`
	Databases LazyValue[map[string]string] `json:"databases,omitempty"`
	Actions   []VimAction                  `json:"actions,omitempty"`
}

type VimAction struct {
	Id   string   `json:"id"`
	Name string   `json:"name"`
	Args []string `json:"args"`
	Cwd  string   `json:"cwd"`
}

type LauncherAction struct {
	Id    string         `json:"id"`
	Tasks []LauncherTask `json:"tasks"`
}

type LauncherTask struct {
	Id           string   `json:"string"`
	Args         []string `json:"args"`
	Cwd          string   `json:"cwd"`
	RunAsService bool     `json:"run_as_service"`
	WorkspaceID  int      `json:"workspace_id"`
}

type Workspace struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	VimConfig VimConfig `json:"vim"`
}

type Context interface {
	FromHome(string) string
	FromEnvDir(string) string
}

type BackupConfig struct {
	GpgKeyring bool
	Secrets    map[string]string
	Data       map[string]string
}

type InitAction struct {
	Args []string `json:"args"`
	Cwd  string   `json:"cwd"`
}

type EnvironmentConfig struct {
	Workspaces        []Workspace
	Actions           []LauncherAction
	Backup            BackupConfig
	Init              []InitAction
	CustomSetupAction func(Context) action.Object
}

type LazyValue[T any] (func() T)

func (l *LazyValue[T]) Resolve() T {
	if *l == nil {
		instance := new(T)
		return *instance
	}
	return (*l)()
}

func (l *LazyValue[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Resolve())
}
