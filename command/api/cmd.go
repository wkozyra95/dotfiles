package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/backup"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/helper"
	"github.com/wkozyra95/dotfiles/api/language"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
	"github.com/wkozyra95/dotfiles/api/sway"
	"github.com/wkozyra95/dotfiles/api/tool"
)

type (
	object   = map[string]interface{}
	endpoint struct {
		name    string
		handler func(context.Context, object) interface{}
	}
)

func SimpleHandler(fn func(context.Context, object) error) func(context.Context, object) interface{} {
	return func(ctx context.Context, input object) interface{} {
		if err := fn(ctx, input); err != nil {
			panic(err)
		}
		return map[string]string{}
	}
}

var endpoints = map[string]endpoint{
	"workspaces:list": {
		name: "workspaces:list",
		handler: func(ctx context.Context, input object) interface{} {
			return ctx.EnvironmentConfig.Workspaces
		},
	},
	"directory:preview": {
		name: "directory:preview",
		handler: func(ctx context.Context, input object) interface{} {
			directoryPreview, err := helper.GetDirectoryPreview(
				getStringField(input, "path"),
				helper.DirectoryPreviewOptions{MaxElements: 20},
			)
			if err != nil {
				panic(err)
			}
			return directoryPreview
		},
	},
	"launch": {
		name: "launch",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return api.AlacrittyRun(input)
		}),
	},
	"node:playground:delete": {
		name: "node:playground:delete",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return language.NodePlaygroundDelete(getStringField(input, "path"))
		}),
	},
	"node:playground:create": {
		name: "node:playground:create",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return language.NodePlaygroundCreate(getStringField(input, "path"))
		}),
	},
	"node:playground:node-shell": {
		name: "node:playground:node-shell",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return language.NodePlaygroundNodeShell(getStringField(input, "path"))
		}),
	},
	"node:playground:zsh-shell": {
		name: "node:playground:zsh-shell",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return language.NodePlaygroundZshShell(getStringField(input, "path"))
		}),
	},
	"node:playground:install": {
		name: "node:playground:install",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return language.NodePlaygroundInstall(getStringField(input, "path"), getStringField(input, "package"))
		}),
	},
	"docker:playground:create": {
		name: "docker:playground:create",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return tool.DockerPlaygroundCreate(getStringField(input, "path"), getStringField(input, "image"))
		}),
	},
	"docker:playground:shell": {
		name: "docker:playground:shell",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return tool.DockerPlaygroundShell(getStringField(input, "path"))
		}),
	},
	"elixir:lsp:install": {
		name: "elixir:lsp:install",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return action.RunSilent(nvim.ElixirLspInstallAction(ctx, true))
		}),
	},
	"terminal:new": {
		name: "terminal:new",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return sway.OpenTerminal(ctx)
		}),
	},
	"backup:zsh_history": {
		name: "backup:zsh_history",
		handler: SimpleHandler(func(ctx context.Context, input object) error {
			return backup.BackupZSHHistory(ctx)
		}),
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
				input = map[string]interface{}{}
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
			result := endpoint.handler(ctx, input)
			serialized, serializeErr := json.Marshal(result)
			if serializeErr != nil {
				panic(serializeErr)
			}
			fmt.Printf("%s\n", string(serialized))
		},
	}
	apiCmd.PersistentFlags().BoolVar(
		&simple, "simple", false,
		"simple api call (api argument is just a name and it's not base64 encoded)",
	)
	rootCmd.AddCommand(apiCmd)
}
