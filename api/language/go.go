package language

import (
	"fmt"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

type goInstallActionArgs struct {
	executable      string
	pkg             string
	shouldReinstall bool
}

func GoInstallAction(executable string, pkg string, shouldReinstall bool) action.Object {
	return goInstallAction(goInstallActionArgs{
		pkg:             pkg,
		executable:      executable,
		shouldReinstall: shouldReinstall,
	})
}

var goInstallAction = action.SimpleActionBuilder[goInstallActionArgs]{
	CreateRun: func(p goInstallActionArgs) func() error {
		return func() error {
			if exec.CommandExists(p.executable) && !p.shouldReinstall {
				return nil
			}
			return exec.Command().WithStdio().Run("go", "install", p.pkg)
		}
	},
	String: func(p goInstallActionArgs) string {
		return fmt.Sprintf("go install %s", p.pkg)
	},
}.Init()
