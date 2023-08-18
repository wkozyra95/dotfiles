package tool

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/manifoldco/promptui"

	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

type Drive struct {
	name       string
	device     string
	partitions []Partition
}

func (d Drive) String() string {
	return fmt.Sprintf(
		"dev: %s (%s)",
		d.name, d.device,
	)
}

func (d Drive) Device() string {
	return d.device
}

func (d Drive) Partitions() []Partition {
	return d.partitions
}

type Partition struct {
	name       string
	device     string
	mountPoint string
}

func (d Partition) MountPoint() string {
	return d.mountPoint
}

func (p Partition) Device() string {
	return p.device
}

func (d Partition) GetPath() string {
	if d.mountPoint == "" {
		panic("unknown mount point, drive was not mounted")
	}
	return d.mountPoint
}

func (d Partition) String() string {
	return fmt.Sprintf(
		"dev: %s (%s), mount: %s",
		d.name, d.device, d.mountPoint,
	)
}

func (d Partition) IsMounted() bool {
	return d.mountPoint != ""
}

func (d *Partition) Mount() error {
	log.Infof("Mounting device %s", d.device)
	mount := fmt.Sprintf("/media/%s", d.name)
	if !file.Exists(mount) {
		if err := exec.Command().WithSudo().WithStdio().Run("mkdir", "-p", mount); err != nil {
			log.Errorf("Unable to create mount point [%v]", err)
			return err
		}
	}
	okPmount, pmountErr := d.mountWithPmount(mount)
	if pmountErr != nil {
		log.Warnf("Command \"pmount\" failed with error %v", pmountErr)
	} else if !okPmount {
		log.Info("Command \"pmount\" not available, falling back to other alternatives")
	} else {
		return nil
	}
	log.Info("tests")
	return d.mountWithMount(mount)
}

func (d *Partition) mountWithMount(mount string) error {
	cmdErr := exec.Command().WithSudo().WithStdio().Run("mount", d.device, mount)
	if cmdErr != nil {
		log.Infof("Mount failed with error [%v]", cmdErr)
		return cmdErr
	}
	d.mountPoint = mount
	return nil
}

func (d *Partition) mountWithPmount(mount string) (bool, error) {
	if !exec.CommandExists("pmount") {
		return false, nil
	}
	cmdErr := exec.Command().WithStdio().Run("pmount", d.device, mount)
	if cmdErr != nil {
		log.Infof("Mount failed with error [%v]", cmdErr)
		return false, cmdErr
	}
	d.mountPoint = mount
	return true, nil
}

func (d *Partition) Umount() error {
	log.Infof("Unmounting device %s", d.device)
	time.Sleep(5 * time.Second)
	okPumount, pumountErr := d.umountWithPumount()
	if pumountErr != nil {
		log.Warnf("Command \"pumount\" failed with error %v", pumountErr)
	} else if !okPumount {
		log.Info("Command \"pumount\" not available, falling bac to other alternatives")
	} else {
		return nil
	}
	if err := d.umountWithUmount(); err != nil {
		_, confirmErr := (&promptui.Prompt{Label: "Umount failed, do you want to continue?", IsConfirm: true}).Run()
		if confirmErr != nil {
			return err
		}
		return nil
	}
	return nil
}

func (d *Partition) umountWithUmount() error {
	cmdErr := exec.Command().WithSudo().WithStdio().Run("umount", d.device)
	if cmdErr != nil {
		log.Warnf("Umount failed with error [%v]", cmdErr)
		return cmdErr
	}
	d.mountPoint = ""
	return nil
}

func (d *Partition) umountWithPumount() (bool, error) {
	if !exec.CommandExists("pumount") {
		return false, nil
	}
	cmdErr := exec.Command().WithStdio().Run("pumount", d.device)
	if cmdErr != nil {
		log.Warnf("Umount failed with error [%v]", cmdErr)
		log.Infof("Retrying in 5 seconds ...")
		return false, cmdErr
	}
	d.mountPoint = ""
	return true, nil
}

func FilterPartitions(fn func(Partition) bool, partitions []Partition) []Partition {
	result := make([]Partition, 0, len(partitions))
	for _, partition := range partitions {
		if fn(partition) {
			result = append(result, partition)
		}
	}
	return result
}

func readDevDirectory() ([]os.FileInfo, error) {
	devDir, devDirErr := os.Open("/dev")
	if devDirErr != nil {
		log.Errorf("Error while reading /dev directory [%v]", devDirErr)
		return nil, devDirErr
	}

	fileList, fileListErr := devDir.Readdir(0)
	if fileListErr != nil {
		log.Errorf("Error while reading /dev directory [%v]", fileListErr)
		return nil, fileListErr
	}
	return fileList, nil
}

func DetectDrives() ([]Drive, error) {
	fileList, fileListErr := readDevDirectory()
	if fileListErr != nil {
		return nil, fileListErr
	}
	return DetectSdxDrives(fileList)
}

func DetectPartitions() ([]Partition, error) {
	fileList, fileListErr := readDevDirectory()
	if fileListErr != nil {
		return nil, fileListErr
	}
	return DetectSdxPartitions(fileList)
}

func DetectSdxPartitions(fileList []os.FileInfo) ([]Partition, error) {
	pattern := regexp.MustCompile("^sd[bcdefg][1-9]$")
	var devList []Partition
	for _, file := range fileList {
		if file.Mode()&os.ModeDevice != 0 &&
			pattern.MatchString(file.Name()) {
			devList = append(devList, Partition{
				name:   file.Name(),
				device: fmt.Sprintf("/dev/%s", file.Name()),
			})
		}
	}
	// check for mount points
	mtabBytes, mtabErr := os.ReadFile("/etc/mtab")
	if mtabErr != nil {
		log.Errorf("Unable to read /etc/mtab, [%v]", mtabErr)
		return nil, mtabErr
	}
	mtab := strings.Split(string(mtabBytes), "\n")
	for i, file := range devList {
		mtabPattern, mtabPatternErr := regexp.Compile(
			fmt.Sprintf("^%s ([^\\s]*) ", file.device),
		)
		if mtabPatternErr != nil {
			panic(fmt.Sprintf("Invalid pattern in mtab regex [%v]", mtabPatternErr))
		}
		for _, mtabEntry := range mtab {
			match := mtabPattern.FindStringSubmatch(mtabEntry)
			if len(match) >= 2 {
				devList[i].mountPoint = match[1]
			}

		}
	}
	return devList, nil
}

func DetectSdxDrives(fileList []os.FileInfo) ([]Drive, error) {
	partitions, partitionsErr := DetectSdxPartitions(fileList)
	if partitionsErr != nil {
		return nil, partitionsErr
	}
	pattern := regexp.MustCompile("^sd[bcdefg]$")
	var devList []Drive
	for _, file := range fileList {
		if file.Mode()&os.ModeDevice != 0 &&
			pattern.MatchString(file.Name()) {
			drive := Drive{
				name:       file.Name(),
				device:     fmt.Sprintf("/dev/%s", file.Name()),
				partitions: []Partition{},
			}
			for _, partition := range partitions {
				if strings.HasPrefix(partition.name, file.Name()) {
					drive.partitions = append(drive.partitions, partition)
				}
			}
			devList = append(devList, drive)
		}
	}
	return devList, nil
}
