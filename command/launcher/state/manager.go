package state

import (
	"fmt"
	"os"
	"time"

	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/persistentstate"
	"github.com/wkozyra95/dotfiles/utils/proc"
)

var log = logger.NamedLogger("launcher:state-manager")

// TODO: move that in some proper directory, either in true tmp or in xdg dirs
const statePath = "/tmp/mycli-state.json"

type processState struct {
	TaskID               string    `json:"taskName"`
	ProcessPid           int       `json:"processPid"`
	ProcessSupervisorPid int       `json:"ProcessSupervisorPid"`
	Time                 time.Time `json:"time"`
}

type errorInfo struct {
	TaskID       string       `json:"taskName"`
	ProcessState processState `json:"processState"`
	ErrorMessage string       `json:"errorMessage"`
	Time         time.Time    `json:"time"`
}

type State struct {
	Processes map[string]processState `json:"processes"`
	Errors    map[string]errorInfo    `json:"errors"`
}

type TaskManager struct {
	peristentState persistentstate.StateManager[State]
}

func ensureDefault(s *State) *State {
	if s == nil {
		s = &State{}
	}
	if s.Errors == nil {
		s.Errors = map[string]errorInfo{}
	}
	if s.Processes == nil {
		s.Processes = map[string]processState{}
	}
	return s
}

func GetTaskManager() TaskManager {
	return TaskManager{
		peristentState: persistentstate.GetStateManager(statePath, "mycli", ensureDefault),
	}
}

func (s *TaskManager) RunGuarded(fun func(*State) error) error {
	return s.peristentState.RunGuarded(func(s *State) error {
		return fun(s)
	})
}

func (s *TaskManager) GetState() (*State, error) {
	state, stateErr := s.peristentState.GetState()
	return state, stateErr
}

// RegisterError ...
func (s *State) RegisterError(taskID string, errMessage string) error {
	s.Errors[taskID] = errorInfo{
		TaskID:       taskID,
		ProcessState: processState{TaskID: taskID, ProcessPid: 0},
		ErrorMessage: errMessage,
		Time:         time.Now(),
	}
	return nil
}

// PrintErrors ,,,
func (s *State) PrintErrors() error {
	s.validateManagedProcesses()
	for _, errEntry := range s.Errors {
		log.Errorf(
			"Task failed (pid: %d, name: %s, time %s)\nError message: %s",
			errEntry.ProcessState.ProcessPid,
			errEntry.ProcessState.TaskID,
			errEntry.Time,
			errEntry.ErrorMessage,
		)
	}
	return nil
}

func (s *State) PrintState() {
	s.validateManagedProcesses()
	for _, process := range s.Processes {
		log.Infof(
			"Task running (pid: %d, ppid %d, name: %s, time %s)",
			process.ProcessPid,
			process.ProcessSupervisorPid,
			process.TaskID,
			process.Time,
		)
	}
}

func (s *State) KillTask(taskID string) error {
	s.validateManagedProcesses()
	for _, processState := range s.Processes {
		if processState.TaskID == taskID {
			if os.Getpid() == processState.ProcessSupervisorPid {
				proc.Term(processState.ProcessPid)
				proc.Term(processState.ProcessPid)
				time.Sleep(time.Second)
			} else {
				log.Debugf("Send term signal %d", processState.ProcessSupervisorPid)
				proc.Term(processState.ProcessSupervisorPid)
				// need some time to remove handler for sigterm
				// TODO: should handle that better
				time.Sleep(time.Millisecond * 10)
				proc.Term(processState.ProcessSupervisorPid)
				time.Sleep(time.Second)
			}
			s.validateManagedProcesses()
			return nil
		}
	}
	return nil
}

func (s *State) IsTaskRunning(taskID string) (bool, error) {
	s.validateManagedProcesses()
	for _, processState := range s.Processes {
		if processState.TaskID == taskID {
			return proc.Exists(processState.ProcessPid), nil
		}
	}
	return false, nil
}

func (s *State) IsSupervisorRunning(taskID string) (bool, error) {
	s.validateManagedProcesses()
	for _, processState := range s.Processes {
		if processState.TaskID == taskID {
			return proc.Exists(processState.ProcessSupervisorPid), nil
		}
	}
	return false, nil
}

func (s *State) RegisterTask(taskID string, pid int) error {
	if !proc.Exists(pid) {
		return s.RegisterError(taskID, fmt.Sprintf("Process with pid %d does not exist", pid))
	}
	delete(s.Errors, taskID)
	s.Processes[taskID] = processState{
		TaskID:               taskID,
		ProcessPid:           pid,
		ProcessSupervisorPid: os.Getpid(),
		Time:                 time.Now(),
	}
	return nil
}

func (s *State) validateManagedProcesses() {
	for taskID, processState := range s.Processes {
		if !proc.Exists(processState.ProcessSupervisorPid) {
			log.Debugf("verified process supervisor %d is down", processState.ProcessSupervisorPid)
			delete(s.Processes, taskID)
		} else if processState.ProcessPid != 0 && !proc.Exists(processState.ProcessPid) {
			log.Debugf("verified process %d is down, but it's managed by a supervisor", processState.ProcessSupervisorPid)
			processState.ProcessPid = 0
			s.Processes[taskID] = processState
		}
	}
}
