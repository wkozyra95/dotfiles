package backup

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/tool/drive"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

var log = logger.NamedLogger("backup")

type KnownEncryptedVolume struct {
	Label      string
	VolumeName string
}

var secretsBackupVolume = KnownEncryptedVolume{
	Label:      "SECRETS_BACKUP",
	VolumeName: "secrets_backup",
}

var dataBackupVolume = KnownEncryptedVolume{
	Label:      "DATA_BACKUP",
	VolumeName: "data_backup",
}

var localDataBackupVolume = KnownEncryptedVolume{
	Label:      "DATA_LOCAL_BACKUP",
	VolumeName: "local_data_backup",
}

var diskConfigs = []KnownEncryptedVolume{
	secretsBackupVolume,
	dataBackupVolume,
	localDataBackupVolume,
}

func filterDrivesByName(name string, devices []drive.StorageDevice) (drive.Partition, error) {
	for _, device := range devices {
		for _, partition := range device.Partitions {
			if partition.Label == name {
				return partition, nil
			}
		}
	}
	return drive.Partition{}, fmt.Errorf("No partition %s", name)
}

func withDataBackup(
	ctx context.Context,
	devices []drive.StorageDevice,
	volume KnownEncryptedVolume,
	pathPrefix string,
	fn func(ctx context.Context, backupDestination string) action.Object,
) error {
	partition, partitionsErr := filterDrivesByName(volume.Label, devices)
	if partitionsErr != nil {
		return partitionsErr
	}

	return withEncryptedDrive(
		ctx,
		partition,
		volume.VolumeName,
		pathPrefix,
		fn,
	)
}

func withEncryptedDrive(
	ctx context.Context,
	partition drive.Partition,
	mountName string,
	pathPrefix string,
	fn func(ctx context.Context, backupDestination string) action.Object,
) error {
	mountPoint := fmt.Sprintf("/mnt/%s", mountName)
	backupDir := fmt.Sprintf("/mnt/%s/%s", mountName, pathPrefix)
	mapperPath := fmt.Sprintf("/dev/mapper/%s", mountName)

	actions := action.List{
		action.WithCondition{
			If:   action.Not(action.PathExists(mapperPath)),
			Then: action.ShellCommand("sudo", "cryptsetup", "luksOpen", partition.PartitionPath, mountName),
		},
		action.ShellCommand("sudo", "mkdir", "-p", mountPoint),
		action.ShellCommand("sudo", "mount", mapperPath, mountPoint),
		action.ShellCommand("sudo", "mkdir", "-p", backupDir),
		action.ShellCommand(
			"sudo", "chown", fmt.Sprintf("%s:%s", ctx.Username, ctx.Username), backupDir,
		),
		fn(ctx, backupDir),
		action.ShellCommand("sudo", "umount", mountPoint),
		action.ShellCommand("sudo", "cryptsetup", "luksClose", mountName),
	}

	if err := action.RunActions(actions); err != nil {
		return err
	}
	return nil
}

func RestoreBackup(ctx context.Context) error {
	devices, devicesErr := drive.GetStorageDevicesList()
	if devicesErr != nil {
		return devicesErr
	}

	secretBackupErr := withDataBackup(
		ctx,
		devices,
		secretsBackupVolume,
		ctx.Environment,
		func(ctx context.Context, backupDestination string) action.Object {
			return action.List{
				action.WithCondition{
					If:   action.Const(ctx.EnvironmentConfig.Backup.GpgKeyring),
					Then: restoreGpgKeyringAction(backupDestination),
				},
				restoreFilesAction(backupDestination, ctx.EnvironmentConfig.Backup.Secrets),
			}
		},
	)
	if secretBackupErr != nil {
		log.Infof(secretBackupErr.Error())
	}
	dataBackupErr := withDataBackup(
		ctx,
		devices,
		dataBackupVolume,
		ctx.Environment,
		func(ctx context.Context, backupDestination string) action.Object {
			return action.List{
				action.WithCondition{
					If:   action.Const(ctx.EnvironmentConfig.Backup.GpgKeyring),
					Then: restoreGpgKeyringAction(backupDestination),
				},
				restoreFilesAction(backupDestination, ctx.EnvironmentConfig.Backup.Data),
			}
		},
	)
	if dataBackupErr != nil {
		log.Infof(dataBackupErr.Error())
	}

	return nil
}

func UpdateBackup(ctx context.Context) error {
	devices, devicesErr := drive.GetStorageDevicesList()
	if devicesErr != nil {
		return devicesErr
	}

	secretBackupErr := withDataBackup(
		ctx,
		devices,
		secretsBackupVolume,
		ctx.Environment,
		func(ctx context.Context, backupDestination string) action.Object {
			return action.List{
				action.WithCondition{
					If:   isBitwardenAuthenticated(),
					Then: backupBitwardenAction(backupDestination),
				},
				action.WithCondition{
					If:   action.Const(ctx.EnvironmentConfig.Backup.GpgKeyring),
					Then: backupGpgKeyringAction(backupDestination),
				},
				backupFilesAction(backupDestination, ctx.EnvironmentConfig.Backup.Secrets),
			}
		},
	)
	if secretBackupErr != nil {
		log.Infof(secretBackupErr.Error())
	}

	dataBackupErr := withDataBackup(
		ctx,
		devices,
		dataBackupVolume,
		ctx.Environment,
		func(ctx context.Context, backupDestination string) action.Object {
			return action.List{
				action.WithCondition{
					If:   isBitwardenAuthenticated(),
					Then: backupBitwardenAction(backupDestination),
				},
				action.WithCondition{
					If:   action.Const(ctx.EnvironmentConfig.Backup.GpgKeyring),
					Then: backupGpgKeyringAction(backupDestination),
				},
				backupFilesAction(backupDestination, ctx.EnvironmentConfig.Backup.Data),
			}
		},
	)
	if dataBackupErr != nil {
		log.Infof(dataBackupErr.Error())
	}

	localDataBackupErr := withDataBackup(
		ctx,
		devices,
		localDataBackupVolume,
		ctx.Environment,
		func(ctx context.Context, backupDestination string) action.Object {
			return action.List{
				action.ShellCommand("mkdir", "-p", backupDestination),
				action.WithCondition{
					If:   isBitwardenAuthenticated(),
					Then: backupBitwardenAction(backupDestination),
				},
				action.WithCondition{
					If:   action.Const(ctx.EnvironmentConfig.Backup.GpgKeyring),
					Then: backupGpgKeyringAction(backupDestination),
				},
				backupFilesAction(backupDestination, ctx.EnvironmentConfig.Backup.Data),
			}
		},
	)
	if localDataBackupErr != nil {
		log.Infof(localDataBackupErr.Error())
	}

	return nil
}

func Connect(ctx context.Context) error {
	devices, devicesErr := drive.GetStorageDevicesList()
	if devicesErr != nil {
		return devicesErr
	}

	knownAndConnectedVolumes := []KnownEncryptedVolume{}
	for _, volume := range diskConfigs {
		_, partitionsErr := filterDrivesByName(volume.Label, devices)
		if partitionsErr != nil {
			continue
		}
		knownAndConnectedVolumes = append(knownAndConnectedVolumes, volume)
	}

	selected, didSelect := prompt.SelectPrompt(
		"Select volume to mount",
		knownAndConnectedVolumes,
		func(volume KnownEncryptedVolume) string {
			return volume.VolumeName
		},
	)
	if !didSelect {
		return errors.New("No drive selected")
	}

	connectionErr := withDataBackup(
		ctx,
		devices,
		selected,
		"",
		func(ctx context.Context, backupDestination string) action.Object {
			return action.Execute(exec.Command().WithCwd(backupDestination).WithStdio(), "bash", "-c", "zsh || true")
		},
	)
	if connectionErr != nil {
		log.Infof(connectionErr.Error())
	}
	return nil
}

func BackupZSHHistory(ctx context.Context) error {
	historyFilePath := path.Join(ctx.Homedir, ".zsh_history")
	historyBackupFilePath := path.Join(ctx.Homedir, ".zsh_history.backup")
	historyFileInfo, statHistoryErr := os.Stat(historyFilePath)
	if statHistoryErr != nil {
		return statHistoryErr
	}
	backupFileInfo, statBackupErr := os.Stat(historyBackupFilePath)
	if statBackupErr != nil && !errors.Is(statBackupErr, os.ErrNotExist) {
		return statBackupErr
	}
	if statBackupErr == nil && historyFileInfo.Size() < backupFileInfo.Size() {
		// TODO: send desktop notification when this happens
		return nil
	}
	return file.Copy(historyFilePath, historyBackupFilePath)
}
