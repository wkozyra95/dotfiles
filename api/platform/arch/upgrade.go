package arch

import (
	"fmt"
	"time"

	a "github.com/wkozyra95/dotfiles/action"
)

func (y Yay) UpgradePackages() error {
	timeString := time.Now().Format("20060102_150405")
	yaySnapshot := "/run/btrfs-root/__snapshot/pre-yay-install"
	timestampSnaphsot := fmt.Sprintf("/run/btrfs-root/__snapshot/root_%s", timeString)
	actions := a.List{
		a.WithCondition{
			If:   a.PathExists(yaySnapshot),
			Then: a.ShellCommand("sudo", "btrfs", "subvolume", "delete", yaySnapshot),
		},
		a.ShellCommand("sudo", "btrfs", "subvolume", "snapshot", "-r", "/", timestampSnaphsot),
		a.ShellCommand(
			"sudo",
			"btrfs",
			"subvolume",
			"snapshot",
			timestampSnaphsot,
			yaySnapshot,
		),
		a.ShellCommand("sudo", "grub-mkconfig", "-o", "/boot/grub/grub.cfg"),
		a.ShellCommand("yay", "-Syu"),
	}
	return a.RunActions(actions)
}
