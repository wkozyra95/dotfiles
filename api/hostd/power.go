package hostd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type powerAction struct {
	// logind method that reports whether this process is allowed to do it
	canMethod string
	// systemctl verb
	verb string
}

// Time between accepting the action and going down, so whoever asked for it
// (HTTP client) still gets an answer.
const powerActionDelay = time.Second

// Suspend returns as soon as the suspend is scheduled.
func Suspend() error {
	return powerAction{canMethod: "CanSuspend", verb: "suspend"}.run()
}

// PowerOff returns as soon as the power off is scheduled.
func PowerOff() error {
	return powerAction{canMethod: "CanPowerOff", verb: "poweroff"}.run()
}

func (a powerAction) run() error {
	if err := a.check(); err != nil {
		return err
	}
	go func() {
		time.Sleep(powerActionDelay)
		log.Infof("systemctl %s", a.verb)
		if out, err := exec.Command("systemctl", a.verb).CombinedOutput(); err != nil {
			log.Errorf("systemctl %s: %v: %s", a.verb, err, strings.TrimSpace(string(out)))
		}
	}()
	return nil
}

// check asks logind up front, because the action itself runs later and can no
// longer report a failure. Anything other than "yes" means polkit would refuse
// (or ask for a password), e.g. because of an inhibitor lock.
func (a powerAction) check() error {
	out, err := exec.Command(
		"busctl", "call", "org.freedesktop.login1", "/org/freedesktop/login1",
		"org.freedesktop.login1.Manager", a.canMethod,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("logind %s: %w: %s", a.canMethod, err, strings.TrimSpace(string(out)))
	}
	if result := strings.TrimSpace(string(out)); result != `s "yes"` {
		return fmt.Errorf("logind %s returned %s", a.canMethod, result)
	}
	return nil
}
