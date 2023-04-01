package drive

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

var potentialNvmeStorageDevicePaths = []string{
	"/dev/nvme0n1",
	"/dev/nvme1n1",
	"/dev/nvme2n1",
	"/dev/nvme3n1",
}

var potentialHddStorageDevicePaths = []string{
	"/dev/sda",
	"/dev/sdb",
	"/dev/sdc",
	"/dev/sdd",
	"/dev/sde",
}

var potentialStorageDevicePaths = append(
	potentialNvmeStorageDevicePaths,
	potentialHddStorageDevicePaths...,
)

type Partition struct {
	Label         string
	PartitionPath string
	Size          int
}

type StorageDevice struct {
	Label      string
	DevicePath string
	Size       int
	Partitions []Partition
}

type SfdiskPartition struct {
	Label         string `json:"name"`
	PartitionPath string `json:"node"`
	Sectors       int    `json:"size"`
	Type          string `json:"type"`
	Uuid          string `json:"uuid"`
}

type SfdiskDeviceInfo struct {
	Paritiontable struct {
		Label      string            `json:"label"`
		Unit       string            `json:"unit"`
		SectorSize int               `json:"sectorsize"`
		DevicePath string            `json:"device"`
		LbaStart   int               `json:"firstlba"`
		LbaEnd     int               `json:"lastlba"`
		Partitions []SfdiskPartition `json:"partitions"`
	} `json:"partitiontable"`
}

func GetPartitionPath(devicePath string, partitionNumber int) string {
	for _, dev := range potentialNvmeStorageDevicePaths {
		if dev == devicePath {
			return fmt.Sprintf("%sp%d", devicePath, partitionNumber)
		}
	}
	for _, dev := range potentialHddStorageDevicePaths {
		if dev == devicePath {
			return fmt.Sprintf("%s%d", devicePath, partitionNumber)
		}
	}
	panic(fmt.Sprintf("Device %s is not supported", devicePath))
}

func GetStorageDevicesList() ([]StorageDevice, error) {
	installedStorageDevicesPaths := make([]StorageDevice, 0, len(potentialStorageDevicePaths))
	for _, devicePath := range potentialStorageDevicePaths {
		if file.Exists(devicePath) {
			device := StorageDevice{DevicePath: devicePath}
			deviceInfo, deviceInfoErr := GetDeviceInfo(devicePath)
			if deviceInfoErr != nil {
				return nil, deviceInfoErr
			}
			device.Size = deviceInfo.Paritiontable.SectorSize * (deviceInfo.Paritiontable.LbaEnd - deviceInfo.Paritiontable.LbaStart)
			device.Label = deviceInfo.Paritiontable.Label
			for _, partition := range deviceInfo.Paritiontable.Partitions {
				device.Partitions = append(device.Partitions, Partition{
					Label:         partition.Label,
					PartitionPath: partition.PartitionPath,
					Size:          deviceInfo.Paritiontable.SectorSize * partition.Sectors,
				})
			}
			installedStorageDevicesPaths = append(installedStorageDevicesPaths, device)
		}
	}
	return installedStorageDevicesPaths, nil
}

func GetDeviceInfo(devicePath string) (SfdiskDeviceInfo, error) {
	var stdout bytes.Buffer
	sfdiskCmdErr := exec.Command().WithBufout(&stdout, &bytes.Buffer{}).Run("sudo", "sfdisk", devicePath, "--json")
	if sfdiskCmdErr != nil {
		return SfdiskDeviceInfo{}, nil
	}
	var deviceInfo SfdiskDeviceInfo
	if err := json.Unmarshal(stdout.Bytes(), &deviceInfo); err != nil {
		return deviceInfo, err
	}
	return deviceInfo, nil
}
