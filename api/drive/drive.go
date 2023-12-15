package drive

import (
	"errors"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/tool/drive"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

func Mount(ctx context.Context) error {
	volumes, readDrivesErr := drive.GetAvailableVolumes()
	if readDrivesErr != nil {
		return readDrivesErr
	}

	mountpoints, mountpointErr := drive.GetMountPointsInfo()
	if mountpointErr != nil {
		return mountpointErr
	}

	filtered := []drive.Volume{}

	for _, volume := range volumes {
		isMounted := false
		for _, mount := range mountpoints {
			if mount.Target == volume.KnownVolume.MountPath {
				isMounted = true
				break
			}
		}
		if !isMounted {
			filtered = append(filtered, volume)
		}
	}

	selected, didSelect := prompt.SelectPrompt(
		"Select volume to mount",
		filtered,
		func(volume drive.Volume) string {
			return volume.Description()
		},
	)
	if !didSelect {
		return errors.New("No drive selected")
	}

	return selected.Mount()
}

func Umount(ctx context.Context) error {
	volumes, readDrivesErr := drive.GetAvailableVolumes()
	if readDrivesErr != nil {
		return readDrivesErr
	}

	mountpoints, mountpointErr := drive.GetMountPointsInfo()
	if mountpointErr != nil {
		return mountpointErr
	}

	filtered := []drive.Volume{}
	for _, volume := range volumes {
		for _, mount := range mountpoints {
			if mount.Target == volume.KnownVolume.MountPath {
				filtered = append(filtered, volume)
			}
		}
	}

	selected, didSelect := prompt.SelectPrompt(
		"Select volume to umount",
		filtered,
		func(volume drive.Volume) string {
			return volume.Description()
		},
	)
	if !didSelect {
		return errors.New("No drive selected")
	}

	if err := selected.Umount(); err != nil {
		return err
	}

	isLUKSOpened := selected.IsLUKSOpened()
	if !isLUKSOpened || !prompt.ConfirmPrompt("Do you want to close this encrypted device?") {
		return nil
	}

	return action.RunActions(selected.CloseLUKSAction(), false)
}
