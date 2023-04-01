//go:build integration
// +build integration

package e2e

import (
	"os"
	"os/exec"
	"testing"
)

func TestToolsInstall(t *testing.T) {
	t.Log("Executing go build")
	cmd := exec.Command("go", "build", "-o", "mycli")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "../../"
	if err := cmd.Run(); err != nil {
		t.Error("Compilation failed with error")
		return
	}
	container := dockerInstance{
		dockerfile: "./test/cmd/test.install.Dockerfile",
		cwd:        "../..",
	}
	container.init()
	if err := container.buildImage(); err != nil {
		t.Errorf("Unable to start docker with error %v", err)
	}
}
