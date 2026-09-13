package nas

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

const (
	User      = "wojtek"
	Host      = "192.168.100.5"
	Mac       = "00:d8:61:7a:45:c4"
	Broadcast = "192.168.100.255"
	Flake     = ".#home-nas"
)

func sshTarget() string {
	return fmt.Sprintf("%s@%s", User, Host)
}

func magicPacket(mac string) ([]byte, error) {
	raw, err := hex.DecodeString(strings.ReplaceAll(mac, ":", ""))
	if err != nil || len(raw) != 6 {
		return nil, fmt.Errorf("invalid MAC address %q", mac)
	}
	packet := make([]byte, 0, 6+16*6)
	packet = append(packet, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff)
	for i := 0; i < 16; i++ {
		packet = append(packet, raw...)
	}
	return packet, nil
}

// Wake sends a Wake-on-LAN magic packet to the NAS as a UDP broadcast. Magic
// packets are LAN broadcasts, so this only works from the same network.
func Wake() error {
	packet, err := magicPacket(Mac)
	if err != nil {
		return err
	}
	conn, err := net.Dial("udp", net.JoinHostPort(Broadcast, "9"))
	if err != nil {
		return fmt.Errorf("open UDP socket: %w", err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		if _, err := conn.Write(packet); err != nil {
			return fmt.Errorf("send magic packet: %w", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

// IsUp reports whether the NAS accepts SSH connections.
func IsUp() bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(Host, "22"), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// WaitUntilUp polls SSH until the NAS answers or the timeout passes.
func WaitUntilUp(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if IsUp() {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

// Suspend puts the NAS into suspend-to-RAM over SSH. It comes back on a
// magic packet (Wake) or the power button.
func Suspend() error {
	return exec.Command().WithStdio().
		Args("ssh", "-t", sshTarget(), "sudo", "systemctl", "suspend").
		Run()
}

// Rebuild builds the NAS configuration from the local dotfiles checkout and
// activates it on the NAS. Nothing is built on the NAS itself.
func Rebuild(dotfilesDir string) error {
	return exec.RunAll(
		exec.Command().WithCwd(dotfilesDir).WithStdio().Args("git", "add", "-A"),
		exec.Command().WithCwd(dotfilesDir).WithStdio().
			Args("nixos-rebuild", "switch", "--flake", Flake, "--target-host", sshTarget(), "--ask-sudo-password"),
	)
}
