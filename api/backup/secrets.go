package backup

import (
	"fmt"
	"path"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

func isBitwardenAuthenticated() action.Condition {
	return action.FuncCond(func() (bool, error) {
		if !exec.CommandExists("bw") {
			return false, nil
		}
		err := exec.Command().Run("bw", "login", "--check")
		return err == nil, nil
	})
}

func backupBitwardenAction(rootDir string) action.Object {
	backupFile := path.Join(rootDir, "bitwarden.json")
	return action.List{
		action.ShellCommand("rm", "-rf", backupFile),
		action.ShellCommand("bw", "export", "--format", "json", "--output", backupFile),
	}
}

func backupGpgKeyringAction(rootDir string) action.Object {
	gpgPath := path.Join(rootDir, "gpg")
	publicKeysPath := path.Join(gpgPath, "gpg_public_keys.asc")
	privateKeysPath := path.Join(gpgPath, "gpg_private_keys.asc")
	trustDbPath := path.Join(gpgPath, "gpg_trustdb.txt")
	return action.List{
		action.ShellCommand("mkdir", "-p", gpgPath),
		action.ShellCommand(
			"bash", "-c",
			fmt.Sprintf("gpg --armor --export > %s", publicKeysPath),
		),
		action.ShellCommand(
			"bash", "-c",
			fmt.Sprintf("gpg --armor --export-secret-keys > %s", privateKeysPath),
		),
		action.ShellCommand(
			"bash", "-c",
			fmt.Sprintf("gpg --export-ownertrust > %s", trustDbPath),
		),
	}
}

func restoreGpgKeyringAction(rootDir string) action.Object {
	gpgPath := path.Join(rootDir, "gpg")
	publicKeysPath := path.Join(gpgPath, "gpg_public_keys.asc")
	privateKeysPath := path.Join(gpgPath, "gpg_private_keys.asc")
	trustDbPath := path.Join(gpgPath, "gpg_trustdb.txt")
	return action.List{
		action.ShellCommand("mkdir", "-p", gpgPath),
		action.WithCondition{
			If:   action.PathExists(publicKeysPath),
			Then: action.ShellCommand("bash", "-c", fmt.Sprintf("gpg --import %s", publicKeysPath)),
		},
		action.WithCondition{
			If:   action.PathExists(privateKeysPath),
			Then: action.ShellCommand("bash", "-c", fmt.Sprintf("gpg --import %s", privateKeysPath)),
		},
		action.WithCondition{
			If:   action.PathExists(trustDbPath),
			Then: action.ShellCommand("bash", "-c", fmt.Sprintf("gpg --import-ownertrust %s", trustDbPath)),
		},
	}
}
