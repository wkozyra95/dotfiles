package installer

import (
	"github.com/manifoldco/promptui"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/system/tool"
)

var log = logger.NamedLogger("installer")

func PrepareDrive() (tool.Drive, error) {
	drives, detectErr := tool.DetectDrives()
	if detectErr != nil {
		log.Error(detectErr)
	}

	names := []string{}
	for _, dst := range drives {
		names = append(names, dst.String())
	}

	index, _, selectErr := (&promptui.Select{
		Label: "Which device do you want to provision",
		Items: names,
	}).Run()
	if selectErr != nil {
		log.Errorf("Prompt select failed [%v]", selectErr)
		return tool.Drive{}, selectErr
	}

	for _, partition := range drives[index].Partitions() {
		if partition.IsMounted() {
			if err := partition.Umount(); err != nil {
				return tool.Drive{}, err
			}
		}
	}
	return drives[index], nil
}
