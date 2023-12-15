package setup

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strings"

	a "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/tool/drive"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

type installTarget struct {
	device        string
	efiPartition  string
	mainPartition string
}

func selectInstallTargetDrive(message string) (installTarget, error) {
	target := installTarget{}
	devices, readDevicesErr := drive.GetStorageDevicesList()
	if readDevicesErr != nil {
		return target, nil
	}

	device, isSelected := prompt.SelectPrompt(
		message,
		devices,
		func(d drive.StorageDevice) string {
			return fmt.Sprintf("%s (size: %d MB)", d.DevicePath, d.Size/(1024*1024))
		},
	)
	if !isSelected {
		return target, fmt.Errorf("No value was selected")
	}
	target.device = device.DevicePath
	target.efiPartition = drive.GetPartitionPath(target.device, 2)
	target.mainPartition = drive.GetPartitionPath(target.device, 3)

	return target, nil
}

func ProvisionArchChroot() error {
	username := "wojtek"
	target, targetErr := selectInstallTargetDrive("Which device do you want to provision")
	if targetErr != nil {
		return targetErr
	}
	fromHome := func(relative string) string {
		return path.Join("/mnt/btrfs-current/home", username, relative)
	}
	luksEnabled := true
	rootPartition := target.mainPartition
	if luksEnabled {
		rootPartition = "/dev/mapper/root"
	}
	maybeLuksUuid := ""
	actions := a.List{
		a.ShellCommand("timedatectl", "set-ntp", "true"),
		a.ShellCommand("sgdisk", "-Z", target.device),
		a.ShellCommand("sgdisk", "-a", "2048", "-o", target.device),
		a.ShellCommand(
			"sgdisk",
			"-n",
			"1::+1M",
			"--typecode=1:ef02",
			"--change-name=1:'BIOS boot partition'",
			target.device,
		), // partition 1 (BIOS Boot Partition)
		a.ShellCommand(
			"sgdisk",
			"-n",
			"2::+300M",
			"--typecode=2:ef00",
			"--change-name=2:'EFI system partition'",
			target.device,
		), // partition 2 (UEFI Boot Partition)
		a.ShellCommand(
			"sgdisk",
			"-n",
			"3::-0",
			"--typecode=3:8300",
			"--change-name=3:'Root'",
			target.device,
		), // partition 3 (Root), default start, remaining
		a.WithCondition{
			If:   a.Not(a.PathExists("/sys/firmware/efi")), // if efi is not supported
			Then: a.ShellCommand("sgdisk", "-A", "1:set:2", target.device),
		},
		a.ShellCommand("mkfs.fat", "-F", "32", "-n", "EFIBOOT", target.efiPartition),
		a.WithCondition{
			If: a.Const(luksEnabled),
			Then: a.List{
				a.ShellCommand("cryptsetup", "-y", "-v", "luksFormat", "--type", "luks1", target.mainPartition),
				a.ShellCommand("cryptsetup", "open", target.mainPartition, "root"),
				a.ShellCommand("mkfs.btrfs", "--label", "BTRFS_ROOT", "--force", rootPartition),
			},
			Else: a.List{
				a.ShellCommand("mkfs.btrfs", "--label", "BTRFS_ROOT", "--force", rootPartition),
			},
		},
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-root"),
		a.ShellCommand(
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,nosuid",
			rootPartition,
			"/mnt/btrfs-root",
		),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-root/__current"),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-root/__snapshot"),
		a.ShellCommand("btrfs", "subvolume", "create", "/mnt/btrfs-root/__current/root"),
		a.ShellCommand("btrfs", "subvolume", "create", "/mnt/btrfs-root/__current/home"),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-current"),
		a.ShellCommand(
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,subvol=__current/root",
			rootPartition,
			"/mnt/btrfs-current",
		),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-current/boot/efi"),
		a.ShellCommand("mount", target.efiPartition, "/mnt/btrfs-current/boot/efi"),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-current/home"),
		a.ShellCommand(
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,nosuid,subvol=__current/home",
			rootPartition,
			"/mnt/btrfs-current/home",
		),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-current/run/btrfs-root"),
		a.ShellCommand(
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,nosuid",
			rootPartition,
			"/mnt/btrfs-current/run/btrfs-root",
		),
		a.ShellCommand(
			"pacstrap",
			"/mnt/btrfs-current",
			"base",
			"base-devel",
			"btrfs-progs",
			"linux",
			"linux-firmware",
			"networkmanager",
			"grub-bios",
			"grub",
			"efibootmgr",
			"os-prober",
		),
		a.WithCondition{
			If: a.Const(luksEnabled),
			Then: a.Scope("read LUKS partition UUID", func() a.Object {
				var stdout bytes.Buffer
				err := exec.Command().
					WithBufout(&stdout, &bytes.Buffer{}).
					Run("blkid", "-s", "UUID", "-o", "value", target.mainPartition)
				if err != nil {
					return a.Err(err)
				}
				maybeLuksUuid = strings.Trim(stdout.String(), " \n")
				return a.List{
					a.ShellCommand(
						"dd",
						"bs=512",
						"count=4",
						"if=/dev/random",
						"of=/mnt/btrfs-current/root/cryptlvm.keyfile",
						"iflag=fullblock",
					),
					a.ShellCommand("chmod", "000", "/mnt/btrfs-current/root/cryptlvm.keyfile"),
					a.ShellCommand(
						"cryptsetup",
						"-v",
						"luksAddKey",
						target.mainPartition,
						"/mnt/btrfs-current/root/cryptlvm.keyfile",
					),
					a.EnsureText(
						"/mnt/btrfs-current/etc/default/grub",
						fmt.Sprintf(
							"GRUB_CMDLINE_LINUX=\"cryptdevice=UUID=%s:root  cryptkey=rootfs:/root/cryptlvm.keyfile\"",
							maybeLuksUuid,
						),
						regexp.MustCompile(".*GRUB_CMDLINE_LINUX=.*"),
					),
					a.EnsureText(
						"/mnt/btrfs-current/etc/default/grub",
						"GRUB_ENABLE_CRYPTODISK=y",
						regexp.MustCompile(".*GRUB_ENABLE_CRYPTODISK.*"),
					),
				}
			}),
		},
		a.EnsureText(
			"/mnt/btrfs-current/root/.bash_history",
			fmt.Sprintf(
				"/home/%s/.dotfiles/bin/mycli tool setup:arch:desktop",
				username,
			),
			nil,
		),
		a.ShellCommand("zsh", "-c", "genfstab -L /mnt/btrfs-current >> /mnt/btrfs-current/etc/fstab"),
		a.ShellCommand("sed", "-i", "-E", "s/,subvolid=[0-9]+//g", "/mnt/btrfs-current/etc/fstab"),
		a.ShellCommand("mkdir", "-p", fromHome(".")),
		a.ShellCommand("cp", "-R", "/root/.dotfiles", fromHome(".dotfiles")),
		a.ShellCommand("cp", "-R", "/root/.ssh", fromHome(".ssh")),
		a.ShellCommand("cp", "-R", "/root/.secrets", fromHome(".secrets")),
		a.ShellCommand("arch-chroot", "/mnt/btrfs-current"),
	}
	return a.RunActions(actions, false)
}

func ProvisionArchChrootForCompanionSystem() error {
	username := "wojtek"
	target, targetErr := selectInstallTargetDrive("Which device do you want to provision")
	if targetErr != nil {
		return targetErr
	}
	fromHome := func(relative string) string {
		return path.Join("/mnt/btrfs-current/home", username, relative)
	}
	envName := prompt.TextPrompt("Environment name")
	if envName == "" {
		return fmt.Errorf("Empty value is not allowed")
	}
	volumePath := fmt.Sprintf("/run/btrfs-root/__%s", envName)
	luksEnabled := true
	rootPartition := target.mainPartition
	if luksEnabled {
		rootPartition = "/dev/mapper/root"
	}
	actions := a.List{
		a.ShellCommand("mkdir", "-p", volumePath),
		a.ShellCommand("btrfs", "subvolume", "create", path.Join(volumePath, "/root")),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-current"),
		a.ShellCommand(
			"mount",
			"-o",
			fmt.Sprintf("defaults,relatime,discard,ssd,nodev,subvol=__%s/root", envName),
			rootPartition,
			"/mnt/btrfs-current",
		),
		a.ShellCommand(
			"pacstrap",
			"/mnt/btrfs-current",
			"base",
			"base-devel",
			"btrfs-progs",
			"linux",
			"linux-firmware",
			"networkmanager",
		),
		a.EnsureText(
			"/mnt/btrfs-current/root/.bash_history",
			fmt.Sprintf(
				"/home/%s/.dotfiles/bin/mycli tool setup:arch:desktop  --stage main",
				username,
			),
			nil,
		),
		a.ShellCommand("zsh", "-c", "genfstab -L /mnt/btrfs-current >> /mnt/btrfs-current/etc/fstab"),
		a.ShellCommand("sed", "-i", "-E", "s/,subvolid=[0-9]+//g", "/mnt/btrfs-current/etc/fstab"),
		a.ShellCommand("mkdir", "-p", fromHome(".")),
		a.ShellCommand("cp", "-R", "/home/wojtek/.dotfiles", fromHome(".dotfiles")),
		a.ShellCommand("cp", "-R", "/home/wojtek/.ssh", fromHome(".ssh")),
		a.ShellCommand("cp", "-R", "/home/wojtek/.secrets", fromHome(".secrets")),
		a.ShellCommand("arch-chroot", "/mnt/btrfs-current"),
	}
	return a.RunActions(actions, false)
}

func ConnectToExistingChrootedEnvironment() error {
	target, targetErr := selectInstallTargetDrive(
		"Which device do you want to prepare for chroot environment",
	)
	if targetErr != nil {
		return targetErr
	}
	luksEnabled := true
	rootPartition := target.mainPartition
	if luksEnabled {
		rootPartition = "/dev/mapper/root"
	}
	actions := a.List{
		a.WithCondition{
			If:   a.Const(luksEnabled),
			Then: a.ShellCommand("cryptsetup", "open", target.mainPartition, "root"),
		},
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-root"),
		a.ShellCommand(
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,nosuid",
			rootPartition,
			"/mnt/btrfs-root",
		),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-current"),
		a.ShellCommand(
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,subvol=__current/root",
			target.mainPartition,
			"/mnt/btrfs-current",
		),
		a.ShellCommand("mount", target.efiPartition, "/mnt/btrfs-current/boot/efi"),
		a.ShellCommand(
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,nosuid,subvol=__current/home",
			target.mainPartition,
			"/mnt/btrfs-current/home",
		),
		a.ShellCommand("mkdir", "-p", "/mnt/btrfs-current/run/btrfs-root"),
		a.ShellCommand(
			"mount",
			"-o",
			"defaults,relatime,discard,ssd,nodev,nosuid",
			target.mainPartition,
			"/mnt/btrfs-current/run/btrfs-root",
		),
		a.ShellCommand("arch-chroot", "/mnt/btrfs-current"),
	}

	return a.RunActions(actions, false)
}
