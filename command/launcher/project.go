package launcher

import (
	"fmt"
	goexec "os/exec"
	"strings"
	"time"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/command/launcher/state"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/proc"
)

type launchJobParams struct {
	jobID   string
	restart bool
}

type launchTaskParams struct {
	taskID  string
	jobID   string
	restart bool
}

type launcher struct {
	actions []env.LauncherAction
	manager state.TaskManager
}

func createLauncher(ctx context.Context) (*launcher, error) {
	manager := state.GetTaskManager()
	return &launcher{actions: ctx.EnvironmentConfig.Actions, manager: manager}, nil
}

func (l *launcher) launchJob(params launchJobParams) error {
	action, actionErr := getAction(l.actions, params.jobID)
	if actionErr != nil {
		return actionErr
	}
	for _, task := range action.Tasks {
		if err := l.launchTask(task, params.jobID, params.restart); err != nil {
			log.Errorf("Task %s failed with error %s", task.Id, err.Error())
			return err
		}
	}
	log.Info("Waiting for tasks to start")
	time.Sleep(time.Second * 5)

	if err := l.manager.RunGuarded(func(s *state.State) error {
		return s.PrintErrors()
	}); err != nil {
		return err
	}

	return nil
}

func (l *launcher) printLauncherState() error {
	if err := l.manager.RunGuarded(func(s *state.State) error {
		log.Info("Currently running tasks:")
		s.PrintState()
		log.Info()
		log.Info("Recent errors from running tasks:")
		return s.PrintErrors()
	}); err != nil {
		return err
	}

	return nil
}

func (l *launcher) launchInternalTask(params launchTaskParams) {
	action, actionErr := getAction(l.actions, params.jobID)
	if actionErr != nil {
		log.Errorf("Failed to resolve a value %v", actionErr)
		time.Sleep(time.Second * 10)
		return
	}
	task, taskErr := getTask(action, params.taskID)
	if taskErr != nil {
		log.Errorf("Failed to resolve a task %v", taskErr)
		time.Sleep(time.Second * 10)
		return
	}
	listenerCleanup := proc.StartDoubleInteruptExitGuard()
	defer listenerCleanup()

	for {
		if err := l.doLaunchInternalTask(task, params.restart); err != nil {
			managerErr := l.manager.RunGuarded(func(s *state.State) error {
				return s.RegisterError(
					params.taskID,
					err.Error(),
				)
			})
			log.Errorf("Task %s failed with error %v", task.Id, err)
			if managerErr != nil {
				log.Errorf("Failed to register an error %v", managerErr)
			}
		}
		time.Sleep(time.Second * 5)
	}
}

func (l *launcher) launchTaskAsService(task env.LauncherTask, jobID string, restart bool) error {
	log.Debugf("Launching service %s", task.Id)
	return l.manager.RunGuarded(func(s *state.State) error {
		isTaskSupervisorRunning, isTaskRunningErr := s.IsSupervisorRunning(task.Id)
		if isTaskRunningErr != nil {
			return isTaskRunningErr
		}
		if !isTaskSupervisorRunning || restart {
			if isTaskSupervisorRunning && restart {
				log.Debug("Supervisor for this task is already running, killing existing process")
				if err := s.KillTask(task.Id); err != nil {
					return err
				}
			}
			cmdStr := []string{"alacritty"}
			if task.WorkspaceID != 0 {
				cmdStr = append(cmdStr, "--class", fmt.Sprintf("workspace%d", task.WorkspaceID))
			}
			cmdStr = append(cmdStr, "--command", "mycli", "launch:internal",
				"--job", jobID,
				"--task", task.Id,
			)

			_, cmdErr := exec.Command().Args(cmdStr...).Start()
			if cmdErr != nil {
				return cmdErr
			}
			return nil
		}
		return nil
	})
}

func (l *launcher) launchTask(task env.LauncherTask, jobID string, restart bool) error {
	if task.RunAsService {
		return l.launchTaskAsService(task, jobID, restart)
	}
	log.Debugf("Launching task %s", task.Id)
	err := exec.Command().WithStdio().WithCwd(task.Cwd).Args(task.Args...).Run()
	if err != nil {
		log.Errorf("Task %s failed with error %s", task.Id, err.Error())
		return err
	}
	return nil
}

func (l *launcher) doLaunchInternalTask(task env.LauncherTask, restart bool) error {
	var cmd *goexec.Cmd
	log.Infof("Starting task %s", task.Id)
	startErr := l.manager.RunGuarded(func(s *state.State) error {
		// If there are other supervisors kill
		cmdInProgress, cmdErr := exec.
			Command().WithStdio().WithCwd(task.Cwd).
			Args(task.Args...).Start()
		if cmdErr != nil {
			log.Errorf("Tried to run invalid command %s", strings.Join(task.Args, " "))
			return cmdErr
		}
		cmd = cmdInProgress
		return s.RegisterTask(task.Id, cmd.Process.Pid)
	})
	if startErr != nil {
		return startErr
	}
	if cmd == nil {
		return nil
	}

	log.Info("Waiting for job to finish")
	if err := cmd.Wait(); err != nil {
		log.Errorf("Task %s failed with error %s", task.Id, err.Error())
		err := l.manager.RunGuarded(func(s *state.State) error {
			return s.RegisterError(task.Id, err.Error())
		})
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}
