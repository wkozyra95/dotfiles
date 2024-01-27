package setup

import (
	"fmt"
	"os/user"

	a "github.com/wkozyra95/dotfiles/action"
)

func InstallNixOS() error {
	if user, userErr := user.Current(); userErr != nil || user.Username != "nixos" {
		panic("This command should be only be run from an install media")
	}
	username := "wojtek"
	target, targetErr := selectInstallTargetDrive("Which device do you want to install NixOS on?")
	if targetErr != nil {
		return targetErr
	}
	actions := a.List{
		// Partition
		a.ShellCommand("sudo", "sgdisk", "-Z", target.device),
		a.ShellCommand("sudo", "sgdisk", "-a", "2048", "-o", target.device),
		a.ShellCommand(
			"sudo",
			"sgdisk",
			"-n",
			"1::+1M",
			"--typecode=1:ef02",
			"--change-name=1:'BIOS boot partition'",
			target.device,
		), // partition 1 (BIOS Boot Partition)
		a.ShellCommand(
			"sudo",
			"sgdisk",
			"-n",
			"2::+300M",
			"--typecode=2:ef00",
			"--change-name=2:'EFI system partition'",
			target.device,
		), // partition 2 (UEFI Boot Partition)
		a.ShellCommand(
			"sudo",
			"sgdisk",
			"-n",
			"3::-0",
			"--typecode=3:8300",
			"--change-name=3:'Root'",
			target.device,
		), // partition 3 (Root), default start, remaining

		// Format
		a.ShellCommand("sudo", "mkfs.fat", "-F", "32", "-n", "EFIBOOT", target.efiPartition),
		a.ShellCommand("sudo", "cryptsetup", "-y", "-v", "luksFormat", "--type", "luks1", target.mainPartition),
		a.ShellCommand("sudo", "cryptsetup", "open", target.mainPartition, "root"),
		a.ShellCommand("sudo", "mkfs.btrfs", "--label", "BTRFS_ROOT", "--force", "/dev/mapper/root"),

		// setup btrfs subvolumes and mount volumes
		a.ShellCommand("sudo", "mkdir", "-p", "/mnt/btrfs-root"),
		a.ShellCommand(
			"sudo",
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,nosuid",
			"/dev/mapper/root",
			"/mnt/btrfs-root",
		),
		a.ShellCommand("sudo", "mkdir", "-p", "/mnt/btrfs-root/__current"),
		a.ShellCommand("sudo", "mkdir", "-p", "/mnt/btrfs-root/__snapshot"),
		a.ShellCommand("sudo", "btrfs", "subvolume", "create", "/mnt/btrfs-root/__current/root"),
		a.ShellCommand("sudo", "btrfs", "subvolume", "create", "/mnt/btrfs-root/__current/home"),
		a.ShellCommand("sudo", "mkdir", "-p", "/mnt/btrfs-current"),
		a.ShellCommand(
			"sudo",
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,subvol=__current/root",
			"/dev/mapper/root",
			"/mnt/btrfs-current",
		),
		a.ShellCommand("sudo", "mkdir", "-p", "/mnt/btrfs-current/boot/efi"),
		a.ShellCommand("sudo", "mount", target.efiPartition, "/mnt/btrfs-current/boot/efi"),
		a.ShellCommand("sudo", "mkdir", "-p", "/mnt/btrfs-current/home"),
		a.ShellCommand(
			"sudo",
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,nosuid,subvol=__current/home",
			"/dev/mapper/root",
			"/mnt/btrfs-current/home",
		),

		// add additionl crypt key
		a.ShellCommand(
			"sudo",
			"dd",
			"bs=512",
			"count=4",
			"if=/dev/random",
			"of=/mnt/btrfs-current/root/cryptlvm.keyfile",
			"iflag=fullblock",
		),
		a.ShellCommand("sudo", "chmod", "000", "/mnt/btrfs-current/root/cryptlvm.keyfile"),
		a.ShellCommand(
			"sudo",
			"cryptsetup",
			"-v",
			"luksAddKey",
			target.mainPartition,
			"/mnt/btrfs-current/root/cryptlvm.keyfile",
		),

		// copy dotfiles
		a.ShellCommand("sudo", "mkdir", "-p", fmt.Sprintf("/mnt/btrfs-current/home/%s", username)),
		a.ShellCommand(
			"sudo",
			"cp",
			"-R",
			"/iso/dotfiles",
			fmt.Sprintf("/mnt/btrfs-current/home/%s/.dotfiles", username),
		),

		a.ShellCommand(
			"echo",
			fmt.Sprintf(
				"sudo nixos-install --root /mnt/btrfs-current --flake /mnt/btrfs-current/home/%s/.dotfiles#home",
				username,
			),
		),
	}
	return a.RunActions(actions, false)
}
