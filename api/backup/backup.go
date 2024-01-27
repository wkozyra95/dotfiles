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
	"github.com/wkozyra95/dotfiles/utils/file"
)

var log = logger.NamedLogger("backup")

type KnownEncryptedVolume struct {
	Label      string
	VolumeName string
}

var (
	secretsBackupMountPath      = "/mnt/secrets_usb_drive_backup"
	externalDataBackupMountPath = "/mnt/external_old_hdd_drive_env_data_backup"
	localDataBackupMountPath    = "/mnt/local_hdd_drive_env_data_backup"
)

func withDataBackup(
	ctx context.Context,
	volume drive.Volume,
	pathPrefix string,
	fn func(ctx context.Context, backupDestination string) action.Object,
) error {
	backupDir := path.Join(volume.MountPath, pathPrefix)
	isLUKSClosed := volume.IsEncrypted() && !volume.IsLUKSOpened()
	actions := action.List{
		volume.MountAction(),
		action.ShellCommand("sudo", "mkdir", "-p", backupDir),
		action.ShellCommand(
			"sudo", "chown", fmt.Sprintf("%s:%s", ctx.Username, ctx.Group), backupDir,
		),
		fn(ctx, backupDir),
		volume.UmountAction(),
		action.WithCondition{
			If:   action.Const(isLUKSClosed),
			Then: volume.CloseLUKSAction(),
		},
	}

	if err := action.RunActions(actions, false); err != nil {
		return err
	}
	return nil
}

func RestoreBackup(ctx context.Context) error {
	volumes, volumesErr := drive.GetAvailableVolumes()
	if volumesErr != nil {
		return volumesErr
	}

	var secretsBackupVolume *drive.Volume
	var externalDataBackupVolume *drive.Volume

	for _, volume := range volumes {
		volume := volume
		if volume.KnownVolume.MountPath == externalDataBackupMountPath {
			externalDataBackupVolume = &volume
		} else if volume.KnownVolume.MountPath == secretsBackupMountPath {
			secretsBackupVolume = &volume
		}
	}

	if secretsBackupVolume != nil {
		secretBackupErr := withDataBackup(
			ctx,
			*secretsBackupVolume,
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
	}

	if externalDataBackupVolume != nil {
		dataBackupErr := withDataBackup(
			ctx,
			*externalDataBackupVolume,
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
	}
	return nil
}

func UpdateBackup(ctx context.Context) error {
	volumes, volumesErr := drive.GetAvailableVolumes()
	if volumesErr != nil {
		return volumesErr
	}

	var secretsBackupVolume *drive.Volume
	var localDataBackupVolume *drive.Volume
	var externalDataBackupVolume *drive.Volume

	for _, volume := range volumes {
		volume := volume
		if volume.KnownVolume.MountPath == externalDataBackupMountPath {
			externalDataBackupVolume = &volume
		} else if volume.KnownVolume.MountPath == localDataBackupMountPath {
			localDataBackupVolume = &volume
		} else if volume.KnownVolume.MountPath == secretsBackupMountPath {
			secretsBackupVolume = &volume
		}
	}

	if secretsBackupVolume != nil {
		secretBackupErr := withDataBackup(
			ctx,
			*secretsBackupVolume,
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
	}

	if externalDataBackupVolume != nil {
		dataBackupErr := withDataBackup(
			ctx,
			*externalDataBackupVolume,
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
	}

	if localDataBackupVolume != nil {
		localDataBackupErr := withDataBackup(
			ctx,
			*localDataBackupVolume,
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
