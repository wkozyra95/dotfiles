package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/helper"
	"github.com/wkozyra95/dotfiles/api/language"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
	"github.com/wkozyra95/dotfiles/api/tool"
	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("api")

type (
	object   = map[string]interface{}
	endpoint struct {
		name    string
		handler func(context.Context, object) interface{}
	}
)

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
		handler: func(ctx context.Context, input object) interface{} {
			if err := api.AlacrittyRun(input); err != nil {
				panic(err)
			}
			return map[string]string{}
		},
	},
	"node:playground:delete": {
		name: "node:playground:delete",
		handler: func(ctx context.Context, input object) interface{} {
			if err := language.NodePlaygroundDelete(getStringField(input, "path")); err != nil {
				panic(err)
			}
			return map[string]string{}
		},
	},
	"node:playground:create": {
		name: "node:playground:create",
		handler: func(ctx context.Context, input object) interface{} {
			if err := language.NodePlaygroundCreate(getStringField(input, "path")); err != nil {
				panic(err)
			}
			return map[string]string{}
		},
	},
	"node:playground:node-shell": {
		name: "node:playground:node-shell",
		handler: func(ctx context.Context, input object) interface{} {
			if err := language.NodePlaygroundNodeShell(getStringField(input, "path")); err != nil {
				panic(err)
			}
			return map[string]string{}
		},
	},
	"node:playground:zsh-shell": {
		name: "node:playground:zsh-shell",
		handler: func(ctx context.Context, input object) interface{} {
			if err := language.NodePlaygroundZshShell(getStringField(input, "path")); err != nil {
				panic(err)
			}
			return map[string]string{}
		},
	},
	"node:playground:install": {
		name: "node:playground:install",
		handler: func(ctx context.Context, input object) interface{} {
			err := language.NodePlaygroundInstall(getStringField(input, "path"), getStringField(input, "package"))
			if err != nil {
				panic(err)
			}
			return map[string]string{}
		},
	},
	"docker:playground:create": {
		name: "docker:playground:create",
		handler: func(ctx context.Context, input object) interface{} {
			if err := tool.DockerPlaygroundCreate(getStringField(input, "path"), getStringField(input, "image")); err != nil {
				panic(err)
			}
			return map[string]string{}
		},
	},
	"docker:playground:shell": {
		name: "docker:playground:shell",
		handler: func(ctx context.Context, input object) interface{} {
			if err := tool.DockerPlaygroundShell(getStringField(input, "path")); err != nil {
				panic(err)
			}
			return map[string]string{}
		},
	},
	"elixir:lsp:install": {
		name: "elixir:lsp:install",
		handler: func(ctx context.Context, o object) interface{} {
			if err := action.RunSilent(nvim.ElixirLspInstallAction(ctx, true)); err != nil {
				panic(err)
			}
			return map[string]string{}
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
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "mycli api",
		Long:  "api used by other tools, all commands return json to stdout",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			decodedInput, decodeErr := base64.StdEncoding.DecodeString(args[0])
			if decodeErr != nil {
				panic(decodeErr)
			}
			input := map[string]interface{}{}
			if err := json.Unmarshal(decodedInput, &input); err != nil {
				panic(err)
			}
			if input == nil || input["name"] == nil {
				panic(fmt.Errorf("missing endpoint name"))
			}
			inputName, inputNameIsString := (input["name"]).(string)
			if !inputNameIsString {
				panic(fmt.Errorf("\"name\" has to be a string"))
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

	rootCmd.AddCommand(apiCmd)
}
