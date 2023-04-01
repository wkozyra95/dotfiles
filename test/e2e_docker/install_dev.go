package e2e

import (
	"os"
	"os/exec"

	"github.com/docker/docker/api/types/mount"
)

func RunInDevTestToolsInstall() {
	log.Info("Executing go build")
	cmd := exec.Command("go", "build", "-o", "mycli")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Error("Compilation failed with error")
		panic(err)
	}
	container := dockerInstance{
		dockerfile:    "./dev.Dockerfile",
		cwd:           "./test/cmd",
		containerName: "system_setup_dev",
		imageName:     "system_setup_dev_img",
		mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   "/home/wojtek/.dotfiles",
				Target:   "/home/test/mounted_tmp_system_setup",
				ReadOnly: true,
			},
		},
	}
	if err := container.start(); err != nil {
		log.Error(err.Error())
		panic(err)
	}
}
