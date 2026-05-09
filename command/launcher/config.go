package launcher

import (
	"fmt"

	"github.com/wkozyra95/dotfiles/env"
)

func getTask(action env.LauncherAction, taskID string) (env.LauncherTask, error) {
	for _, task := range action.Tasks {
		if task.ID == taskID {
			return task, nil
		}
	}
	return env.LauncherTask{}, fmt.Errorf("no action named %s", taskID)
}

func getAction(actions []env.LauncherAction, actionID string) (env.LauncherAction, error) {
	for _, action := range actions {
		if action.ID == actionID {
			return action, nil
		}
	}
	return env.LauncherAction{}, fmt.Errorf("no action named %s", actionID)
}
