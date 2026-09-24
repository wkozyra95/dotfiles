// Package hostd implements what the hostd server does on the host: power
// management, file transfer directory, the registry of push tokens of mobile
// devices (delivered to by api/mobile) and the token. Everything related to
// HTTP lives in the server subpackage.
package hostd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("hostd")

type Status struct {
	Hostname string
	Uptime   time.Duration
}

func ReadStatus() (Status, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return Status{}, err
	}
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return Status{}, err
	}
	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return Status{}, fmt.Errorf("unexpected /proc/uptime content %q", content)
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return Status{}, err
	}
	return Status{Hostname: hostname, Uptime: time.Duration(seconds * float64(time.Second))}, nil
}
