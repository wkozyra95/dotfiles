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
	return err == nil
}

// Cmd ...
type Cmd struct {
	innerCmd *exec.Cmd
	stdio    bool
	bufout   io.Writer
	buferr   io.Writer
	sudo     bool
	cwd      string
	args     []string
}

// Command ...
func Command() *Cmd {
	cmd := &exec.Cmd{}
	return &Cmd{cmd, false, nil, nil, false, "", nil}
}

// WithEnv ...
func (c *Cmd) WithEnv(envs ...string) *Cmd {
	if c.innerCmd.Env == nil {
		c.innerCmd.Env = os.Environ()
	}
	c.innerCmd.Env = append(c.innerCmd.Env, envs...)
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

func (c *Cmd) Args(args ...string) *Cmd {
	c.args = append([]string{}, args...)
	return c
}

// Start ...
func (c *Cmd) Start() (*exec.Cmd, error) {
	c.prepare()
	if err := c.innerCmd.Start(); err != nil {
		return c.innerCmd, fmt.Errorf("Command %v failed with error [%s]", c.innerCmd.Args, err.Error())
	}
	return c.innerCmd, nil
}

// Run ...
func (c *Cmd) Run() error {
	c.prepare()
	if c.innerCmd.Stdout != nil {
		if err := c.innerCmd.Run(); err != nil {
			return fmt.Errorf("Command %v failed with error [%s]", c.innerCmd.Args, err.Error())
		}
	} else {
		return c.runWithoutStdio()
	}
	return nil
}

func RunAll(cmds ...*Cmd) error {
	for _, cmd := range cmds {
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cmd) prepare() {
	if c.sudo {
		c.innerCmd.Path = "sudo"
		c.innerCmd.Args = append([]string{"sudo"}, c.args...)
	} else {
		c.innerCmd.Path = c.args[0]
		c.innerCmd.Args = c.args
	}
	if filepath.Base(c.innerCmd.Path) == c.innerCmd.Path {
		if lp, err := exec.LookPath(c.innerCmd.Path); err != nil {
			// return err
		} else {
			c.innerCmd.Path = lp
		}
	}
	c.innerCmd.Dir = c.cwd
	log.Debugf("Call [%s]", strings.Join(c.innerCmd.Args, " "))
	if c.stdio {
		c.innerCmd.Stdout = os.Stdout
		c.innerCmd.Stderr = os.Stderr
		c.innerCmd.Stdin = os.Stdin
	} else if c.bufout != nil && c.buferr != nil {
		c.innerCmd.Stdout = c.bufout
		c.innerCmd.Stderr = c.buferr
		c.innerCmd.Stdin = os.Stdin
	}
}

func (c *Cmd) runWithoutStdio() error {
	cmd := c.innerCmd
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
		log.Error(buffer.String())
		return cmdErr
	}
	return nil
}
