package launcher

import (
	"fmt"

	"github.com/wkozyra95/dotfiles/env"
)

type task struct {
	Cmd          string `json:"cmd"`
	Cwd          string `json:"cwd"`
	RunAsService bool   `json:"runAsService"`
	IsCritical   bool   `json:"isCritical"`
	Stdio        string `json:"stdio"`
	X11Tag       string `json:"x11Tag"`
}

type config struct {
	Jobs  map[string][]string `json:"jobs"`
	Tasks map[string]task     `json:"tasks"`
}

func getTask(action env.LauncherAction, taskID string) (env.LauncherTask, error) {
	for _, task := range action.Tasks {
		if task.Id == taskID {
			return task, nil
		}
	}
	return env.LauncherTask{}, fmt.Errorf("No action named %s", taskID)
}

func getAction(actions []env.LauncherAction, actionID string) (env.LauncherAction, error) {
	for _, action := range actions {
		if action.Id == actionID {
			return action, nil
		}
	}
	return env.LauncherAction{}, fmt.Errorf("No action named %s", actionID)
}
