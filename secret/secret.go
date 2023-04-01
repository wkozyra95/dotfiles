package secret

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"

	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

const GIT_CRYPT_MAGIC_STRING = "\x00GITCRYPT"

var log = logger.NamedLogger("exec")

type WifiConfig struct {
	Ssid     string `json:"ssid"`
	Password string `json:"password"`
}

type Secrets struct {
	Wifi struct {
		HomeOrange WifiConfig `json:"home_orange"`
		HomeUpc    WifiConfig `json:"home_upc"`
	} `json:"wifi"`
}

type fileEncryptedError string

func (f fileEncryptedError) Error() string {
	return string(f)
}

func ReadSecret(homedir string) (Secrets, error) {
	secretsFile := path.Join(homedir, "./.dotfiles/secret/secrets.json")
	file, readErr := ioutil.ReadFile(secretsFile)
	if readErr != nil {
		return Secrets{}, readErr
	}
	isEncrypted := strings.HasPrefix(string(file), GIT_CRYPT_MAGIC_STRING)
	if isEncrypted {
		return Secrets{}, fileEncryptedError("File is encrypted")
	}
	secrets := Secrets{}
	if err := json.Unmarshal(file, &secrets); err != nil {
		return Secrets{}, err
	}
	return secrets, nil
}

func BestEffortReadSecrets(homedir string) *Secrets {
	secrets, readSecretsErr := ReadSecret(homedir)
	if readSecretsErr == nil {
		return &secrets
	}
	if _, isEncrypted := readSecretsErr.(fileEncryptedError); !isEncrypted {
		log.Errorf("Failed to read secrets %v", readSecretsErr)
		return nil
	}
	if !prompt.ConfirmPrompt("You secrets are encrypted, do you want to try unlocking the repo") {
		log.Error("Skipping secrets read")
		return nil
	}
	if !exec.CommandExists("git-crypt") {
		log.Error("Skipping secrets read, git-crypt is not installed")
		return nil
	}
	if err := exec.Command().WithStdio().Run("git-crypt", "unlock", path.Join(homedir, ".secrets/dotfiles.key")); err != nil {
		log.Errorf("Skipping secrets read, unlock failed with %s", err.Error())
		return nil
	}
	secrets, readSecretsErr = ReadSecret(homedir)
	if readSecretsErr != nil {
		log.Errorf("Skipping secrets read, failed to read secrets %s", readSecretsErr.Error())
		return nil
	}
	return &secrets
}
