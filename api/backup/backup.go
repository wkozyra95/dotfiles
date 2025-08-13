package backup

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/tool/drive"
	"github.com/wkozyra95/dotfiles/utils/exec"
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

func cmd() *exec.Cmd {
	return exec.Command().WithStdio()
}

func sudo() *exec.Cmd {
	return exec.Command().WithStdio().WithSudo()
}

func runFnWithBackupVolume(
	ctx context.Context,
	volume drive.Volume,
	pathPrefix string,
	fn func(backupDestination string) error,
) (resultErr error) {
	backupDir := path.Join(volume.MountPath, pathPrefix)
	isLUKSClosed := volume.IsEncrypted() && !volume.IsLUKSOpened()

	if err := volume.Mount(); err != nil {
		return err
	}
	defer func() {
		if err := volume.Umount(); err != nil {
			if resultErr == nil {
				resultErr = err
			}
			return
		}
		if isLUKSClosed {
			err := volume.CloseLUKS()
			if resultErr == nil {
				resultErr = err
			}
		}
	}()

	cmdsErr := exec.RunAll(
		sudo().Args("mkdir", "-p", backupDir),
		sudo().Args("chown", fmt.Sprintf("%s:%s", ctx.Username, ctx.Group), backupDir),
	)
	if cmdsErr != nil {
		return cmdsErr
	}
	return fn(backupDir)
}

func RestoreBackup(ctx context.Context) error {
	volumes, volumesErr := drive.GetAvailableVolumes()
	if volumesErr != nil {
		return volumesErr
	}

	var secretsBackupVolume *drive.Volume
	var localDataBackupVolume *drive.Volume

	for _, volume := range volumes {
		volume := volume
		if volume.KnownVolume.MountPath == localDataBackupMountPath {
			localDataBackupVolume = &volume
		} else if volume.KnownVolume.MountPath == secretsBackupMountPath {
			secretsBackupVolume = &volume
		}
	}

	restoreSecretsFn := func(backupDestination string) error {
		if ctx.EnvironmentConfig.Backup.GpgKeyring {
			if err := restoreGpgKeyring(backupDestination); err != nil {
				return err
			}
		}
		return restoreFiles(backupDestination, ctx.EnvironmentConfig.Backup.Secrets)
	}
	restoreData := func(backupDestination string) error {
		if ctx.EnvironmentConfig.Backup.GpgKeyring {
			if err := restoreGpgKeyring(backupDestination); err != nil {
				return err
			}
		}
		return restoreFiles(backupDestination, ctx.EnvironmentConfig.Backup.Data)
	}

	if secretsBackupVolume != nil {
		secretBackupErr := runFnWithBackupVolume(
			ctx,
			*secretsBackupVolume,
			ctx.Environment,
			restoreSecretsFn,
		)
		if secretBackupErr != nil {
			log.Infof(secretBackupErr.Error())
		}
	}

	if localDataBackupVolume != nil {
		dataBackupErr := runFnWithBackupVolume(
			ctx,
			*localDataBackupVolume,
			ctx.Environment,
			restoreData,
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

	secretsBackupFn := func(backupDestination string) error {
		if isBitwardenAuthenticated() {
			if err := backupBitwarden(backupDestination); err != nil {
				return err
			}
		}
		if ctx.EnvironmentConfig.Backup.GpgKeyring {
			if err := backupGpgKeyring(backupDestination); err != nil {
				return err
			}
		}
		return backupFiles(backupDestination, ctx.EnvironmentConfig.Backup.Secrets)
	}
	dataBackupFn := func(backupDestination string) error {
		if isBitwardenAuthenticated() {
			if err := backupBitwarden(backupDestination); err != nil {
				return err
			}
		}
		if ctx.EnvironmentConfig.Backup.GpgKeyring {
			if err := backupGpgKeyring(backupDestination); err != nil {
				return err
			}
		}
		return backupFiles(backupDestination, ctx.EnvironmentConfig.Backup.Data)
	}

	if secretsBackupVolume != nil {
		secretBackupErr := runFnWithBackupVolume(
			ctx,
			*secretsBackupVolume,
			ctx.Environment,
			secretsBackupFn,
		)
		if secretBackupErr != nil {
			log.Infof(secretBackupErr.Error())
		}
	}

	if externalDataBackupVolume != nil {
		dataBackupErr := runFnWithBackupVolume(
			ctx,
			*externalDataBackupVolume,
			ctx.Environment,
			dataBackupFn,
		)
		if dataBackupErr != nil {
			log.Infof(dataBackupErr.Error())
		}
	}

	if localDataBackupVolume != nil {
		localDataBackupErr := runFnWithBackupVolume(
			ctx,
			*localDataBackupVolume,
			ctx.Environment,
			dataBackupFn,
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
