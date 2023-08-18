package bitwarden

import (
	"bytes"
	"encoding/json"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

func IsLoggedIn() (bool, error) {
	var stdout, stderr bytes.Buffer
	if err := exec.Command().WithBufout(&stdout, &stderr).Run("bw", "status"); err != nil {
		return false, err
	}
	var output map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return false, err
	}
	if output["status"] == "authenticated" {
		return true, nil
	}
	return false, nil
}

func InteractiveLogin() error {
	return exec.Command().WithStdio().Run("bw", "login")
}

func EnsureLoggedIn() error {
	isLoggedIn, err := IsLoggedIn()
	if err != nil {
		return err
	}
	if isLoggedIn {
		return nil
	}
	if err := InteractiveLogin(); err != nil {
		return err
	}
	return nil
}
