package setup

import (
	"fmt"
	"os/user"

	"github.com/wkozyra95/dotfiles/utils/exec"
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

	partitionErr := exec.RunAll(
		sudo().Args("sgdisk", "-Z", target.device),
		sudo().Args("sgdisk", "-a", "2048", "-o", target.device),
		// partition 1 (BIOS Boot Partition)
		sudo().Args(
			"sgdisk",
			"-n",
			"1::+1M",
			"--typecode=1:ef02",
			"--change-name=1:'BIOS boot partition'",
			target.device,
		),
		// partition 2 (UEFI Boot Partition)
		sudo().Args(
			"sgdisk",
			"-n",
			"2::+300M",
			"--typecode=2:ef00",
			"--change-name=2:'EFI system partition'",
			target.device,
		),
		// partition 3 (Root), default start, remaining
		sudo().Args(
			"sudo",
			"sgdisk",
			"-n",
			"3::-0",
			"--typecode=3:8300",
			"--change-name=3:'Root'",
			target.device,
		),
	)
	if partitionErr != nil {
		return partitionErr
	}

	formatPartitionsErr := exec.RunAll(
		sudo().Args("mkfs.fat", "-F", "32", "-n", "EFIBOOT", target.efiPartition),
		sudo().Args("mkfs.setup", "-y", "-v", "luksFormat", "--type", "luks1", target.mainPartition),
		sudo().Args("mkfs.setup", "open", target.mainPartition, "root"),
		sudo().Args("mkfs.btrfs", "--label", "BTRFS_ROOT", "--force", "/dev/mapper/root"),
	)
	if formatPartitionsErr != nil {
		return formatPartitionsErr
	}

	setupBtrfsSubvolumesErr := exec.RunAll(
		sudo().Args("mkdir", "-p", "/mnt/btrfs-root"),
		sudo().Args(
			"mount",
			"-o", "defaults,relatime,discard,ssd,nodev,nosuid",
			"/dev/mapper/root", "/mnt/btrfs-root",
		),
		sudo().Args("mkdir", "-p", "/mnt/btrfs-root/__current"),
		sudo().Args("mkdir", "-p", "/mnt/btrfs-root/__snapshot"),
		sudo().Args("btrfs", "subvolume", "create", "/mnt/btrfs-root/__current/root"),
		sudo().Args("btrfs", "subvolume", "create", "/mnt/btrfs-root/__current/home"),
		sudo().Args("mkdir", "-p", "/mnt/btrfs-current"),
		sudo().Args(
			"mount",
			"-o", "defaults,relatime,discard,ssd,nodev,subvol=__current/root",
			"/dev/mapper/root", "/mnt/btrfs-current",
		),
		sudo().Args("mkdir", "-p", "/mnt/btrfs-current/boot/efi"),
		sudo().Args("mount", target.efiPartition, "/mnt/btrfs-current/boot/efi"),
		sudo().Args("mkdir", "-p", "/mnt/btrfs-current/home"),
		sudo().Args(
			"mount",
			"-o", "defaults,relatime,discard,ssd,nodev,nosuid,subvol=__current/home",
			"/dev/mapper/root", "/mnt/btrfs-current/home",
		),
	)
	if setupBtrfsSubvolumesErr != nil {
		return setupBtrfsSubvolumesErr
	}

	generateCryptKeyErr := exec.RunAll(
		sudo().Args(
			"dd",
			"bs=512",
			"count=4",
			"if=/dev/random",
			"of=/mnt/btrfs-current/root/cryptlvm.keyfile",
			"iflag=fullblock",
		),
		sudo().Args("chmod", "000", "/mnt/btrfs-current/root/cryptlvm.keyfile"),
		sudo().Args(
			"cryptsetup",
			"-v",
			"luksAddKey",
			target.mainPartition,
			"/mnt/btrfs-current/root/cryptlvm.keyfile",
		),
	)
	if generateCryptKeyErr != nil {
		return generateCryptKeyErr
	}

	copyFilesErr := exec.RunAll(
		sudo().Args("mkdir", "-p", fmt.Sprintf("/mnt/btrfs-current/home/%s", username)),
		sudo().Args(
			"cp", "-R", "/iso/dotfiles",
			fmt.Sprintf("/mnt/btrfs-current/home/%s/.dotfiles", username),
		),
	)
	if copyFilesErr != nil {
		return copyFilesErr
	}

	log.Infof(
		"sudo nixos-install --root /mnt/btrfs-current --flake /mnt/btrfs-current/home/%s/.dotfiles#home",
		username,
	)
	return nil
}
