package setup

import (
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

func cmd() *exec.Cmd {
	return exec.Command().WithStdio()
}

func sudo() *exec.Cmd {
	return exec.Command().WithStdio().WithSudo()
}

var log = logger.NamedLogger("setup")
