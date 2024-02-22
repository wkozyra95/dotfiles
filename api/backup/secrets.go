package backup

import (
	"fmt"
	"path"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

func isBitwardenAuthenticated() bool {
	if !exec.CommandExists("bw") {
		return false
	}
	err := exec.Command().Args("bw", "login", "--check").Run()
	return err == nil
}

func backupBitwarden(rootDir string) error {
	backupFile := path.Join(rootDir, "bitwarden.json")
	return exec.RunAll(
		cmd().Args("rm", "-rf", backupFile),
		cmd().Args("bw", "export", "--format", "json", "--output", backupFile),
	)
}

func backupGpgKeyring(rootDir string) error {
	gpgPath := path.Join(rootDir, "gpg")
	publicKeysPath := path.Join(gpgPath, "gpg_public_keys.asc")
	privateKeysPath := path.Join(gpgPath, "gpg_private_keys.asc")
	trustDbPath := path.Join(gpgPath, "gpg_trustdb.txt")
	return exec.RunAll(
		cmd().Args("mkdir", "-p", gpgPath),
		cmd().Args("bash", "-c", fmt.Sprintf("gpg --armor --export > %s", publicKeysPath)),
		cmd().Args("bash", "-c", fmt.Sprintf("gpg --armor --export-secret-keys > %s", privateKeysPath)),
		cmd().Args("bash", "-c", fmt.Sprintf("gpg --export-ownertrust > %s", trustDbPath)),
	)
}

func restoreGpgKeyring(rootDir string) error {
	gpgPath := path.Join(rootDir, "gpg")
	publicKeysPath := path.Join(gpgPath, "gpg_public_keys.asc")
	privateKeysPath := path.Join(gpgPath, "gpg_private_keys.asc")
	trustDbPath := path.Join(gpgPath, "gpg_trustdb.txt")

	if err := cmd().Args("mkdir", "-p", gpgPath).Run(); err != nil {
		return err
	}
	cmds := []*exec.Cmd{
		cmd().Args("mkdir", "-p", gpgPath),
	}
	if file.Exists(publicKeysPath) {
		cmds = append(cmds,
			cmd().Args("bash", "-c", fmt.Sprintf("gpg --import %s", publicKeysPath)),
		)
	}
	if file.Exists(privateKeysPath) {
		cmds = append(cmds,
			cmd().Args("bash", "-c", fmt.Sprintf("gpg --import %s", privateKeysPath)),
		)
	}
	if file.Exists(trustDbPath) {
		cmds = append(cmds,
			cmd().Args("bash", "-c", fmt.Sprintf("gpg --import-ownertrust %s", trustDbPath)),
		)
	}
	return exec.RunAll(cmds...)
}
