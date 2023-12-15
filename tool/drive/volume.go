package drive

import (
	"fmt"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/utils/file"
)

type KnownVolumePartition struct {
	Label string
	// optional
	LUKSDeviceMapperName string // name of the file in /dev/mapper
	Description          string
}

type KnownVolume struct {
	MountPath string
	Partition KnownVolumePartition
	// optional
	BtrfsSubvolumePath string
}

type Volume struct {
	KnownVolume
	StorageDevice
	Partition
}

var externalOldHDDDrive = KnownVolumePartition{
	Label:                "EXTERNAL_HDD_OLD",
	LUKSDeviceMapperName: "external_data_hdd_old",
	Description:          "external HDD drive (very old)",
}

var externalHDDDrive = KnownVolumePartition{
	Label:       "EXTERNAL_HDD_MULTIMEDIA",
	Description: "external HDD drive 2TB",
}

var externalSSDDrive = KnownVolumePartition{
	Label:                "EXTERNAL_SSD",
	Description:          "external SSD",
	LUKSDeviceMapperName: "external_ssd",
}

var usbDrive = KnownVolumePartition{
	Label:                "SECRETS_BACKUP",
	LUKSDeviceMapperName: "usb_drive_secrets_backup",
	Description:          "usb drive with secrets",
}

var localHDDDrive = KnownVolumePartition{
	Label:                "LOCAL_DATA_HDD",
	LUKSDeviceMapperName: "local_data_hdd",
	Description:          "internal HDD drive",
}

// Mounted using /etc/fstab
// var _localDataSSDDrive = KnownVolumePartition{
// 	Label:                "LOCAL_DATA_SSD",
// 	LUKSDeviceMapperName: "local_data_ssd",
// 	Description:          "internal SDD drive",
// }

// --------------------------------
// Volumes
// --------------------------------

var usbDriveSecretsBackup = KnownVolume{
	MountPath: "/mnt/secrets_usb_drive_backup",
	Partition: usbDrive,
}

var localHDDDriveRoot = KnownVolume{
	MountPath: "/mnt/local_hdd_drive_root",
	Partition: localHDDDrive,
}

var localHDDDriveEnvDataBackup = KnownVolume{
	MountPath:          "/mnt/local_hdd_drive_env_data_backup",
	Partition:          localHDDDrive,
	BtrfsSubvolumePath: "__env_data_backup",
}

var externalOldHDDDriveEnvDataBackup = KnownVolume{
	MountPath: "/mnt/external_old_hdd_drive_env_data_backup",
	Partition: externalOldHDDDrive,
}

var externalSSDRoot = KnownVolume{
	MountPath: "/mnt/external_ssd",
	Partition: externalSSDDrive,
}

var externalHDDMultimedia = KnownVolume{
	MountPath: "/mnt/multimedia",
	Partition: externalHDDDrive,
}

var knownVolumes = []KnownVolume{
	externalSSDRoot,
	externalHDDMultimedia,
	externalOldHDDDriveEnvDataBackup,
	localHDDDriveEnvDataBackup,
	localHDDDriveRoot,
	usbDriveSecretsBackup,
}

func GetAvailableVolumes() ([]Volume, error) {
	devices, devicesErr := GetStorageDevicesList()
	if devicesErr != nil {
		return nil, devicesErr
	}

	volumes := []Volume{}
	for _, device := range devices {
		for _, partition := range device.Partitions {
			for _, volume := range knownVolumes {
				if volume.Partition.Label == partition.Label {
					volumes = append(volumes, Volume{
						KnownVolume:   volume,
						StorageDevice: device,
						Partition:     partition,
					})
				}
			}
		}
	}

	return volumes, nil
}

func (v *Volume) Description() string {
	maybeBtrfs := ""
	if v.KnownVolume.BtrfsSubvolumePath != "" {
		maybeBtrfs = fmt.Sprintf(" subvol=%s", v.KnownVolume.BtrfsSubvolumePath)
	}
	return fmt.Sprintf(
		"[%s]%s - mountpoint=%s",
		v.KnownVolume.Partition.Description, maybeBtrfs, v.MountPath,
	)
}

func (v *Volume) Mount() error {
	return action.RunActions(v.MountAction(), false)
}

func (v *Volume) Umount() error {
	return action.RunActions(v.UmountAction(), false)
}

func (v *Volume) MountAction() action.Object {
	openAction, mountDevicePath := v.maybeLUKSOpen()
	return action.List{
		openAction,
		action.ShellCommand("sudo", "mkdir", "-p", v.MountPath),
		v.mountAction(mountDevicePath),
	}
}

func (v *Volume) UmountAction() action.Object {
	return action.ShellCommand("sudo", "umount", v.MountPath)
}

func (v *Volume) IsEncrypted() bool {
	return v.KnownVolume.Partition.LUKSDeviceMapperName != ""
}

func (v *Volume) IsLUKSOpened() bool {
	if v.KnownVolume.Partition.LUKSDeviceMapperName == "" {
		return false
	}
	mapperDevicePath := fmt.Sprintf("/dev/mapper/%s", v.KnownVolume.Partition.LUKSDeviceMapperName)
	return file.Exists(mapperDevicePath)
}

func (v *Volume) CloseLUKSAction() action.Object {
	return action.ShellCommand("sudo", "cryptsetup", "close", v.KnownVolume.Partition.LUKSDeviceMapperName)
}

func (v *Volume) maybeLUKSOpen() (action.Object, string) {
	if v.KnownVolume.Partition.LUKSDeviceMapperName == "" {
		return action.Nop(), v.Partition.PartitionPath
	}
	mapperDevicePath := fmt.Sprintf("/dev/mapper/%s", v.KnownVolume.Partition.LUKSDeviceMapperName)
	return action.WithCondition{
		If: action.Not(action.PathExists(mapperDevicePath)),
		Then: action.ShellCommand(
			"sudo",
			"cryptsetup",
			"open",
			v.Partition.PartitionPath,
			v.KnownVolume.Partition.LUKSDeviceMapperName,
		),
	}, mapperDevicePath
}

func (v *Volume) mountAction(mountDevicePath string) action.Object {
	if v.BtrfsSubvolumePath == "" {
		return action.ShellCommand("sudo", "mount", mountDevicePath, v.MountPath)
	}

	return action.ShellCommand(
		"sudo",
		"mount",
		"-o",
		fmt.Sprintf("defaults,relatime,discard,ssd,nodev,subvol=%s", v.BtrfsSubvolumePath),
		mountDevicePath,
		v.MountPath,
	)
}
