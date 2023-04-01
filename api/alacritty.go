package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

type AlacrittyConfig struct {
	Command string
	Args    []string
	Cwd     string
}

func AlacrittyCall(params AlacrittyConfig) error {
	config := map[string]interface{}{
		"name":    "launch",
		"command": params.Command,
		"args":    params.Args,
		"cwd":     params.Cwd,
	}
	rawJson, jsonMarshalErr := json.Marshal(config)
	if jsonMarshalErr != nil {
		return jsonMarshalErr
	}
	baseEncodedString := base64.StdEncoding.EncodeToString(rawJson)
	return exec.Command().Run("alacritty", "-e", "mycli", "api", baseEncodedString)
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
			fmt.Printf(err.Error())
		}
		fmt.Println("Press the Enter Key to continue")
		fmt.Scanln()
	}
}
