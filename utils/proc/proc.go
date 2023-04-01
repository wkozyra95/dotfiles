package proc

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("utils:proc")

// Exists checks if process specified by pid exists
func Exists(pid int) bool {
	process, processErr := os.FindProcess(pid)
	if processErr != nil {
		panic(fmt.Sprintf("should happen only on windows %s", processErr.Error()))
	}
	singalErr := process.Signal(syscall.Signal(0))
	if singalErr != nil {
		log.Debug(singalErr.Error())
	}
	return singalErr == nil
}

// Term ...
func Term(pid int) {
	process, processErr := os.FindProcess(pid)
	if processErr != nil {
		panic(fmt.Sprintf("should happen only on windows %s", processErr.Error()))
	}
	process.Signal(syscall.SIGTERM)
}

type DoubleInteruptExitGuard interface {
	Cleanup()
}

func StartDoubleInteruptExitGuard() func() {
	sig := make(chan os.Signal, 1)
	closed := false
	go func() {
		for {
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
			<-sig
			if closed {
				return
			}
			signal.Stop(sig)
			log.Info("Press Ctrl+C again to kill this process")
			time.Sleep(time.Second * 2)
		}
	}()
	return func() {
		close(sig)
		closed = true
	}
}
