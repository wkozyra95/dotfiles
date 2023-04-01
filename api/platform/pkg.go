package platform

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/platform/arch"
	"github.com/wkozyra95/dotfiles/api/platform/macos"
	"github.com/wkozyra95/dotfiles/api/platform/ubuntu"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

var log = logger.NamedLogger("platform")

func GetPackageManager() (api.PackageInstaller, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := exec.Command().WithBufout(&stdout, &stderr).Run("uname", "-s"); err != nil {
		log.Error(stderr.String())
		panic(err)
	}
	osType := strings.Trim(stdout.String(), " \n\t\r")
	if osType == "Linux" {
		if exec.CommandExists("pacman") {
			return arch.Yay{}, nil
		}
		if exec.CommandExists("apt-get") {
			return ubuntu.Apt{}, nil
		}
	} else if osType == "Darwin" {
		return macos.Brew{}, nil
	}
	return nil, fmt.Errorf("unknown platform (%s)", osType)
}
