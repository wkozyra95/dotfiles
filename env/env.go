package env

import (
	"encoding/json"

	"github.com/wkozyra95/dotfiles/action"
)

const (
	Workspace1  int = 1
	Workspace2  int = 2
	Workspace3  int = 3
	Workspace4  int = 4
	Workspace5  int = 5
	Workspace6  int = 6
	Workspace7  int = 7
	Workspace8  int = 8
	Workspace9  int = 9
	Workspace10 int = 10
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

type DockerEnvSpec struct {
	Name           string `json:"name"`
	ImageName      string `json:"image-name"`
	DockerfilePath string `json:"dockerfile-path"`
	ContainerName  string `json:"container-name-prefix"`
}

type EnvironmentConfig struct {
	Workspaces        []Workspace
	Actions           []LauncherAction
	Backup            BackupConfig
	Init              []InitAction
	CustomSetupAction func(Context) action.Object
	DockerEnvsSpec    []DockerEnvSpec
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
