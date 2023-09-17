package notify

import (
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

var log = logger.NamedLogger("notify")

func Notify(title string, body string) {
	err := exec.Command().WithStdio().Run("notify-send", title, body)
	if err != nil {
		log.Errorf("Failed to send notify %v", err)
	}
}
