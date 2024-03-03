package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/backup"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/helper"
	"github.com/wkozyra95/dotfiles/api/language"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
	"github.com/wkozyra95/dotfiles/api/sway"
	"github.com/wkozyra95/dotfiles/api/tool"
	"github.com/wkozyra95/dotfiles/utils/notify"
	"github.com/wkozyra95/dotfiles/utils/term"
)

type (
	object   = map[string]any
	endpoint struct {
		name             string
		interactiveShell bool
		handler          handleFunc
	}
	handleFunc = func(context.Context, object) (any, error)
)

var endpoints = map[string]endpoint{
	"workspaces:list": {
		name: "workspaces:list",
		handler: func(ctx context.Context, input object) (any, error) {
			return ctx.EnvironmentConfig.Workspaces, nil
		},
	},
	"directory:preview": {
		name: "directory:preview",
		handler: func(ctx context.Context, input object) (any, error) {
			directoryPreview, err := helper.GetDirectoryPreview(
				getStringField(input, "path"),
				helper.DirectoryPreviewOptions{MaxElements: 20},
			)
			if err != nil {
				panic(err)
			}
			return directoryPreview, nil
		},
	},
	"launch": {
		name:             "launch",
		interactiveShell: true,
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, api.AlacrittyRun(input)
		},
	},
	"node:playground:delete": {
		name: "node:playground:delete",
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, language.NodePlaygroundDelete(getStringField(input, "path"))
		},
	},
	"node:playground:create": {
		name: "node:playground:create",
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, language.NodePlaygroundCreate(getStringField(input, "path"))
		},
	},
	"node:playground:node-shell": {
		name:             "node:playground:node-shell",
		interactiveShell: true,
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, language.NodePlaygroundNodeShell(getStringField(input, "path"))
		},
	},
	"node:playground:zsh-shell": {
		name:             "node:playground:zsh-shell",
		interactiveShell: true,
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, language.NodePlaygroundZshShell(getStringField(input, "path"))
		},
	},
	"node:playground:install": {
		name: "node:playground:install",
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, language.NodePlaygroundInstall(getStringField(input, "path"), getStringField(input, "package"))
		},
	},
	"docker:playground:create": {
		name: "docker:playground:create",
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, tool.DockerPlaygroundCreate(getStringField(input, "path"), getStringField(input, "image"))
		},
	},
	"docker:playground:shell": {
		name:             "docker:playground:shell",
		interactiveShell: true,
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, tool.DockerPlaygroundShell(getStringField(input, "path"))
		},
	},
	"elixir:lsp:install": {
		name: "elixir:lsp:install",
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, nvim.InstallElixirLSP(ctx, true)
		},
	},
	"terminal:new": {
		name: "terminal:new",
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, sway.OpenTerminal(ctx)
		},
	},
	"backup:zsh_history": {
		name: "backup:zsh_history",
		handler: func(ctx context.Context, input object) (any, error) {
			return nil, backup.BackupZSHHistory(ctx)
		},
	},
}

func getStringField(o object, field string) string {
	anyValue, exists := o[field]
	if !exists {
		panic(fmt.Errorf("field %s does not exists", field))
	}
	stringValue, isString := anyValue.(string)
	if !isString {
		panic(errors.New("field %s has to be a string"))
	}
	return stringValue
}

// RegisterCmds ...
func RegisterCmds(rootCmd *cobra.Command) {
	simple := false
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "mycli api",
		Long:  "api used by other tools, all commands return json to stdout",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var input map[string]any
			var inputName string
			if simple {
				inputName = args[0]
			} else {
				decodedInput, decodeErr := base64.StdEncoding.DecodeString(args[0])
				if decodeErr != nil {
					panic(decodeErr)
				}
				input = map[string]any{}
				if err := json.Unmarshal(decodedInput, &input); err != nil {
					panic(err)
				}
				if input == nil || input["name"] == nil {
					panic(fmt.Errorf("missing endpoint name"))
				}
				name, inputNameIsString := (input["name"]).(string)
				if !inputNameIsString {
					panic(fmt.Errorf("\"name\" has to be a string"))
				}
				inputName = name
			}

			endpoint, endpointExists := endpoints[inputName]
			if !endpointExists {
				panic(fmt.Errorf("endpoint %s does not exists", inputName))
			}
			ctx := context.CreateContext()
			if !endpoint.interactiveShell {
				handleWithStdioRedirect(endpoint, ctx, input)
			} else {
				handleWithRealStdio(endpoint, ctx, input)
			}
		},
	}
	apiCmd.PersistentFlags().BoolVar(
		&simple, "simple", false,
		"simple api call (api argument is just a name and it's not base64 encoded)",
	)
	rootCmd.AddCommand(apiCmd)
}

func handleWithRealStdio(e endpoint, ctx context.Context, input map[string]any) {
	result, commandErr := e.handler(ctx, input)
	if commandErr != nil {
		notify.Notify("Command failed", commandErr.Error())
		result = map[string]any{}
	}
	serialized, serializeErr := json.Marshal(result)
	if serializeErr != nil {
		notify.Notify("Failed to serialize api response", serializeErr.Error())
		panic(serializeErr)
	}
	fmt.Println(string(serialized))
}

func commandCallID(e endpoint) string {
	return fmt.Sprintf("api:%s-%d:%d", e.name, time.Now().Nanosecond(), rand.Intn(100))
}

func handleWithStdioRedirect(e endpoint, ctx context.Context, input map[string]any) {
	logfile := fmt.Sprintf("/tmp/mycli/log:%s", commandCallID(e))
	mkdirErr := os.MkdirAll(path.Dir(logfile), os.ModePerm)
	if mkdirErr != nil {
		panic(mkdirErr)
	}
	redirects, redirectErr := term.RedirectStdioToFile(logfile)
	if redirectErr != nil {
		notify.Notify("Failed to redirect", redirectErr.Error())
		panic(redirectErr)
	}
	result, commandErr := e.handler(ctx, input)
	if commandErr != nil {
		notify.Notify("Command failed", commandErr.Error())
		result = map[string]any{}
	}
	serialized, serializeErr := json.Marshal(result)
	if serializeErr != nil {
		notify.Notify("Failed to serialize api response", serializeErr.Error())
		panic(serializeErr)
	}
	fmt.Fprintln(redirects.Stdout.RealStream(), string(serialized))
	defer redirects.Cleanup()
}
