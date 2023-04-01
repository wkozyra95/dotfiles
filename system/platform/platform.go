package platform

import (
	"bytes"
	"errors"
	"regexp"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

type platform string

var (
	Debian = platform("debian")
	Arch   = platform("arch")
	MacOS  = platform("macos")
)

var (
	archRegexp   = regexp.MustCompile("(?i).*arch.*")
	debianRegexp = regexp.MustCompile("(?i).*debian.*")
	macosRegexp  = regexp.MustCompile("(?i).*darwin.*")
)

func Detect() (platform, error) {
	var cmdOut bytes.Buffer
	if err := exec.Command().WithBufout(&cmdOut, &bytes.Buffer{}).Run("uname", "-a"); err != nil {
		return platform(""), err
	}

	if archRegexp.Match(cmdOut.Bytes()) {
		return Arch, nil
	}
	if debianRegexp.Match(cmdOut.Bytes()) {
		return Debian, nil
	}
	if macosRegexp.Match(cmdOut.Bytes()) {
		return MacOS, nil
	}
	return platform(""), errors.New("Unsupported platform")
}
