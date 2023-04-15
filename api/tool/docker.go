package tool

import (
	"fmt"
	"os"
	"path"

	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

func DockerPlaygroundCreate(playgroundPath string, image string) error {
	if file.Exists(playgroundPath) {
		return nil
	}
	if err := exec.Command().Run("mkdir", "-p", playgroundPath); err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(playgroundPath, "Dockerfile"), []byte(fmt.Sprintf("FROM %s", image)), 0o644); err != nil {
		return err
	}
	return nil
}

func DockerPlaygroundShell(playgroundPath string) error {
	return api.AlacrittyCall(
		api.AlacrittyConfig{
			Command:     "zsh",
			Args:        []string{"-c", "docker build -t test . && docker run -it test"},
			Cwd:         playgroundPath,
			ShouldRetry: true,
		},
	)
}
