package tool

import (
	"errors"
	"fmt"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

type DD struct {
	Input       string
	Output      string
	ChunkSizeKB int
	Status      string
}

func (d DD) Run() error {
	args := []string{}
	if d.Input == "" {
		return errors.New("input can't be empty")
	}
	args = append(args, fmt.Sprintf("if=%s", d.Input))
	if d.Output == "" {
		return errors.New("input can't be empty")
	}
	args = append(args, fmt.Sprintf("of=%s", d.Output))

	if d.ChunkSizeKB != 0 {
		args = append(args, fmt.Sprintf("bs=%dK", d.ChunkSizeKB))
	}
	if d.Status != "" {
		args = append(args, fmt.Sprintf("status=%s", d.Status))
		args = append(args, "conv=fsync", "oflag=direct")
	}

	return exec.Command().WithStdio().WithSudo().Run("dd", args...)
}
