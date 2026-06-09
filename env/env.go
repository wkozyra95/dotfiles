package env

import (
	"encoding/json"

	"github.com/wkozyra95/dotfiles/api/sway"
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

// SessionTask is a single terminal launched by a SessionTemplate.
type SessionTask struct {
	ID   string   `json:"id"`
	Args []string `json:"args"`
	Cwd  string   `json:"cwd"`
	// Slot selects the target workspace within the unit: 0 = primary (focused),
	// 1 = partner.
	Slot int `json:"slot"`
}

// SessionTemplate is a project blueprint launched into the current workspace
// pair by the session manager.
type SessionTemplate struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// DefaultName, when non-empty, is the project name claimed automatically for
	// a freshly launched session: if no session of that name already exists it is
	// used without prompting; only on a collision is the user asked to type one.
	// Empty means always prompt (e.g. templates whose name must be unique per
	// launch, such as worktree templates that derive a branch from it).
	DefaultName string        `json:"default_name,omitempty"`
	Tasks       []SessionTask `json:"tasks"`
	// Prepare, when set, runs at launch with the prompted project name and
	// returns the working directory the tasks should run in (overriding each
	// task's Cwd). Used e.g. to create a git worktree for the project.
	Prepare func(projectName string) (string, error) `json:"-"`
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
	Workspaces []Workspace
	// SessionTemplates is resolved lazily (only when the session picker opens),
	// so dynamic templates can run git/IO without paying for it on every
	// mycli invocation. May be nil.
	SessionTemplates  func() []SessionTemplate
	Backup            BackupConfig
	Init              []InitAction
	CustomSetupAction func(Context) error
	DockerEnvsSpec    []DockerEnvSpec
	SwayHandlers      []sway.Handler
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
