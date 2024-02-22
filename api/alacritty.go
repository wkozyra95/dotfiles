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
	_, err := exec.Command().Args("alacritty", "-e", "mycli", "api", baseEncodedString).Start()
	return err
}

// Entrypoint of a process that was started in a new alacritty window created using AlacrittyCall
func AlacrittyRun(params map[string]interface{}) error {
	args := []string{}
	if params["args"] != nil {
		args = make([]string, len(params["args"].([]interface{})))
		for i, arg := range params["args"].([]interface{}) {
			args[i] = arg.(string)
		}
	}
	for {
		cmd := exec.Command().
			WithStdio().
			WithCwd(params["cwd"].(string)).
			Args(append([]string{params["command"].(string)}, args...)...)
		if err := cmd.Run(); err != nil {
			fmt.Println(err.Error())
		}
		if shouldRetry, isBool := (params["should_retry"]).(bool); isBool && !shouldRetry {
			return nil
		}
		fmt.Println("Press the Enter Key to continue")
		fmt.Scanln()
	}
}
