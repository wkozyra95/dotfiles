package setup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	a "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/system/tool"
	"github.com/wkozyra95/dotfiles/tool/drive"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

func selectPortableInstallMedia() (string, error) {
	devices, readDevicesErr := drive.GetStorageDevicesList()
	if readDevicesErr != nil {
		return "", nil
	}
	filteredDevices := []drive.StorageDevice{}
	for _, device := range devices {
		if device.Label == "dos" {
			filteredDevices = append(filteredDevices, device)
		}
	}

	device, isSelected := prompt.SelectPrompt(
		"Which device do you want to provision",
		filteredDevices,
		func(d drive.StorageDevice) string {
			return fmt.Sprintf("%s (size: %d MB)", d.DevicePath, d.Size/(1024*1024))
		},
	)
	if !isSelected {
		return "", fmt.Errorf("No value was selected")
	}

	return device.DevicePath, nil
}

func ProvisionUsbArchInstaller(ctx context.Context) error {
	target, targetErr := selectPortableInstallMedia()
	if targetErr != nil {
		return targetErr
	}

	workingdir := path.Join(os.TempDir(), "arch-bootable-workingdir")
	if !file.Exists("/usr/share/archiso/configs/releng") {
		if err := exec.Command().WithStdio().Run("yay", "-S", "archiso"); err != nil {
			return err
		}
	}

	fromRootDir := func(relative string) string {
		return path.Join(workingdir, "airootfs", "root", relative)
	}

	actions := a.List{
		a.ShellCommand("sudo", "rm", "-rf", workingdir),
		a.ShellCommand("cp", "-R", "/usr/share/archiso/configs/releng", workingdir),
		a.ShellCommand("mkdir", "-p", fromRootDir(".")),
		a.ShellCommand("cp", "-R", ctx.FromHome(".dotfiles"), fromRootDir(".dotfiles")),
		a.ShellCommand("cp", "-R", ctx.FromHome(".ssh"), fromRootDir(".ssh")),
		a.ShellCommand("cp", "-R", ctx.FromHome(".secrets"), fromRootDir(".secrets")),
		a.EnsureText(path.Join(workingdir, "packages.x86_64"), "networkmanager", nil),
		a.EnsureText(path.Join(workingdir, "packages.x86_64"), "go", nil),
		a.EnsureText(path.Join(workingdir, "packages.x86_64"), "make", nil),
		a.EnsureText(
			fromRootDir(".zsh_history"),
			": 1601313583:0;chmod +x ./.dotfiles/bin/mycli && ./.dotfiles/bin/mycli tool setup:arch:chroot",
			nil,
		),
		a.ShellCommand(
			"sudo", "mkarchiso", "-v",
			"-w", path.Join(workingdir, "tmpdir"),
			"-o", path.Join(workingdir, "out"),
			workingdir,
		),
		a.Func("Write iso to drive", func() error {
			files, fileErr := ioutil.ReadDir(path.Join(workingdir, "out"))
			if fileErr != nil {
				return fileErr
			}

			outputIso := path.Join(workingdir, "out", files[0].Name())

			if !prompt.ConfirmPrompt(fmt.Sprintf("Do you want to copy files to %s device", target)) {
				return fmt.Errorf("Aborting ...")
			}
			return tool.DD{
				Input:       outputIso,
				Output:      target,
				ChunkSizeKB: 4 * 1024,
				Status:      "progress",
			}.Run()
		}),
		a.ShellCommand("sudo", "rm", "-rf", workingdir),
	}

	return a.Run(actions)
}
