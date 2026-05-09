package env

import (
	"encoding/json"
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

type VimFiletypeConfig struct {
	IndentSize int `json:"indent_size"`
}

type VimConfig struct {
	GoEfm          map[string]any               `json:"go_efm,omitempty"`
	CmakeEfm       map[string]any               `json:"cmake_efm,omitempty"`
	FiletypeConfig map[string]VimFiletypeConfig `json:"filetype_config,omitempty"`
	JsonlsSchemas  []JSONSchema                 `json:"json_schemas,omitempty"`
	YamllsSchemas  []JSONSchema                 `json:"yml_schemas,omitempty"`
	Eslint         *bool                        `json:"eslint,omitempty"`
	Databases      LazyValue[map[string]string] `json:"databases,omitempty"`
	Actions        []VimAction                  `json:"actions,omitempty"`
}

type VimAction struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Args []string `json:"args"`
	Cwd  string   `json:"cwd"`
}

type JSONSchema struct {
	FileMatch []string `json:"fileMatch"`
	URL       string   `json:"url"`
}

type LauncherAction struct {
	ID    string         `json:"id"`
	Tasks []LauncherTask `json:"tasks"`
}

type LauncherTask struct {
	ID           string   `json:"string"`
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

type SwayWorkspaceBinding struct {
	Workspace string
	Output    string
}

// SwayWorkspacePair defines two workspaces (each pinned to its own output)
// that should be focused together. When the user switches to either side,
// the listener will switch the other side's output to the paired workspace.
type SwayWorkspacePair struct {
	A SwayWorkspaceBinding
	B SwayWorkspaceBinding
}

type EnvironmentConfig struct {
	Workspaces         []Workspace
	Actions            []LauncherAction
	Backup             BackupConfig
	Init               []InitAction
	CustomSetupAction  func(Context) error
	DockerEnvsSpec     []DockerEnvSpec
	SwayWorkspacePairs []SwayWorkspacePair
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
