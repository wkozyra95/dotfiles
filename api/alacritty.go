package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

type AlacrittyConfig struct {
	Command     string
	Args        []string
	Cwd         string
	ShouldRetry bool
	Async       bool
}

func AlacrittyCall(params AlacrittyConfig) error {
	config := map[string]interface{}{
		"name":         "launch",
		"command":      params.Command,
		"args":         params.Args,
		"cwd":          params.Cwd,
		"should_retry": params.ShouldRetry,
	}
	rawJson, jsonMarshalErr := json.Marshal(config)
	if jsonMarshalErr != nil {
		return jsonMarshalErr
	}
	baseEncodedString := base64.StdEncoding.EncodeToString(rawJson)
	if params.Async {
		_, err := exec.Command().Start("alacritty", "-e", "mycli", "api", baseEncodedString)
		return err
	} else {
		return exec.Command().Run("alacritty", "-e", "mycli", "api", baseEncodedString)
	}
}

func AlacrittyRun(params map[string]interface{}) error {
	args := []string{}
	if params["args"] != nil {
		args = make([]string, len(params["args"].([]interface{})))
		for i, arg := range params["args"].([]interface{}) {
			args[i] = arg.(string)
		}
	}
	for {
		if err := exec.Command().WithStdio().WithCwd(params["cwd"].(string)).Run(params["command"].(string), args...); err != nil {
			fmt.Println(err.Error())
		}
		if shouldRetry, isBool := (params["should_retry"]).(bool); isBool && !shouldRetry {
			return nil
		}
		fmt.Println("Press the Enter Key to continue")
		fmt.Scanln()
	}
}
