package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("exec")

// LookPath ...
func LookPath(cmd string) string {
	str, err := exec.LookPath(cmd)
	if err != nil {
		return ""
	}
	return str
}

// CommandExists ...
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return false
	}
	return true
}

// Cmd ...
type Cmd struct {
	*exec.Cmd
	stdio  bool
	bufout io.Writer
	buferr io.Writer
	sudo   bool
	cwd    string
}

// Command ...
func Command() *Cmd {
	cmd := &exec.Cmd{}
	return &Cmd{cmd, false, nil, nil, false, ""}
}

// WithEnv ...
func (c *Cmd) WithEnv(envs ...string) *Cmd {
	if c.Env == nil {
		c.Env = os.Environ()
	}
	c.Env = append(c.Env, envs...)
	return c
}

// WithStdio ...
func (c *Cmd) WithStdio() *Cmd {
	c.stdio = true
	return c
}

// WithBufout ...
func (c *Cmd) WithBufout(stdout io.Writer, stderr io.Writer) *Cmd {
	c.bufout = stdout
	c.buferr = stderr
	return c
}

// WithCwd ...
func (c *Cmd) WithCwd(cwd string) *Cmd {
	c.cwd = cwd
	return c
}

// WithSudo ...
func (c *Cmd) WithSudo() *Cmd {
	c.sudo = true
	return c
}

// Start ...
func (c *Cmd) Start(cmdName string, args ...string) (*exec.Cmd, error) {
	c.prepare(cmdName, args...)
	if err := c.Cmd.Start(); err != nil {
		return c.Cmd, fmt.Errorf("Command %v failed with error [%s]", c.Cmd.Args, err.Error())
	}
	return c.Cmd, nil
}

// Run ...
func (c *Cmd) Run(cmdName string, args ...string) error {
	c.prepare(cmdName, args...)
	if c.Stdout != nil {
		if err := c.Cmd.Run(); err != nil {
			return fmt.Errorf("Command %v failed with error [%s]", c.Cmd.Args, err.Error())
		}
	} else {
		return c.runWithoutStdio()
	}
	return nil
}

func (c *Cmd) prepare(cmdName string, args ...string) {
	if c.sudo {
		c.Path = "sudo"
		c.Args = append([]string{"sudo", cmdName}, args...)
	} else {
		c.Path = cmdName
		c.Args = append([]string{cmdName}, args...)
	}
	if filepath.Base(c.Path) == c.Path {
		if lp, err := exec.LookPath(c.Path); err != nil {
			// return err
		} else {
			c.Path = lp
		}
	}
	c.Cmd.Dir = c.cwd
	log.Debugf("Call [%s]", strings.Join(c.Args, " "))
	if c.stdio {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
	} else if c.bufout != nil && c.buferr != nil {
		c.Stdout = c.bufout
		c.Stderr = c.buferr
	}
}

func (c *Cmd) runWithoutStdio() error {
	cmd := c.Cmd
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		log.Errorf("Command [%s %v] failed", cmd.Path, cmd.Args)
		lines := strings.Split(string(output), "\n")
		var buffer bytes.Buffer
		buffer.WriteString("Failed command error")
		for _, line := range lines {
			buffer.WriteString("\n\t\t")
			buffer.WriteString(line)
		}
		log.Errorf(buffer.String())
		return cmdErr
	}
	return nil
}
