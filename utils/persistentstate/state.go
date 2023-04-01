package persistentstate

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/juju/mutex"

	"github.com/davecgh/go-spew/spew"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/file"
)

var log = logger.NamedLogger("util:persistentstate")

type stateManager[State any] struct {
	statePath  string
	mutextSpec mutex.Spec
	initFunc   func(*State) *State
}

// StateManager ...
type StateManager[State any] interface {
	RunGuarded(func(*State) error) error
	GetState() (*State, error)
}

type clock struct {
	delay time.Duration
}

func (f *clock) After(t time.Duration) <-chan time.Time {
	return time.After(t)
}

func (f *clock) Now() time.Time {
	return time.Now()
}

func GetStateManager[State any](statePath string, name string, initFn func(*State) *State) StateManager[State] {
	return &stateManager[State]{
		statePath: statePath,
		mutextSpec: mutex.Spec{
			Name:    name,
			Clock:   &clock{},
			Delay:   time.Second,
			Timeout: 0,
			Cancel:  make(chan struct{}),
		},
		initFunc: initFn,
	}
}

func (s *stateManager[State]) RunGuarded(fun func(*State) error) error {
	mutexLock, mutexErr := mutex.Acquire(s.mutextSpec)
	if mutexErr != nil {
		return mutexErr
	}
	defer mutexLock.Release()
	state, readErr := s.read()
	if readErr != nil {
		return readErr
	}
	state = s.initFunc(state)
	log.Tracef("Read state %s", spew.Sdump(state))
	if err := fun(state); err != nil {
		return err
	}
	log.Tracef("Write state %s", spew.Sdump(state))
	return s.write(state)
}

func (s *stateManager[State]) GetState() (*State, error) {
	state, readErr := s.read()
	if readErr != nil {
		emptyState := new(State)
		return emptyState, readErr
	}
	return state, nil
}

func (s *stateManager[State]) read() (*State, error) {
	if file.Exists(s.statePath) {
		rawFile, readErr := ioutil.ReadFile(s.statePath)
		if readErr != nil {
			return nil, readErr
		}
		state := new(State)
		if err := json.Unmarshal(rawFile, state); err != nil {
			return nil, err
		}
		return state, nil
	}
	return new(State), nil
}

func (s *stateManager[State]) write(state *State) error {
	mkdirErr := os.MkdirAll(path.Dir(s.statePath), os.ModePerm)
	if mkdirErr != nil {
		return mkdirErr
	}
	rawState, marshalErr := json.Marshal(&state)
	if marshalErr != nil {
		return marshalErr
	}
	if err := os.WriteFile(s.statePath, rawState, os.ModePerm); err != nil {
		return err
	}
	return nil
}
