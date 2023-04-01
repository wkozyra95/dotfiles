package tool

import (
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/system/tool"
)

var driveMountCmd = &cobra.Command{
	Use:   "mount",
	Short: "mount drive",
	Run: func(cmd *cobra.Command, args []string) {
		drives, detectErr := tool.DetectPartitions()
		if detectErr != nil {
			log.Error(detectErr)
		}

		mounted := tool.FilterPartitions(func(d tool.Partition) bool { return !d.IsMounted() }, drives)

		names := []string{}
		for _, dst := range mounted {
			names = append(names, dst.String())
		}

		index, _, selectErr := (&promptui.Select{
			Label: "Which device you want to mount",
			Items: names,
		}).Run()
		if selectErr != nil {
			log.Errorf("Prompt select failed [%v]", selectErr)
			return
		}
		if err := mounted[index].Mount(); err != nil {
			log.Error(err)
		}
	},
}

var driveUmountCmd = &cobra.Command{
	Use:   "umount",
	Short: "umount drive",
	Run: func(cmd *cobra.Command, args []string) {
		drives, detectErr := tool.DetectPartitions()
		if detectErr != nil {
			log.Error(detectErr)
		}

		mounted := tool.FilterPartitions(func(d tool.Partition) bool { return d.IsMounted() }, drives)

		names := []string{}
		for _, dst := range mounted {
			names = append(names, dst.String())
		}

		index, _, selectErr := (&promptui.Select{
			Label: "Which device you want to umount",
			Items: names,
		}).Run()
		if selectErr != nil {
			log.Errorf("Prompt select failed [%v]", selectErr)
			return
		}
		if err := mounted[index].Umount(); err != nil {
			log.Error(err)
		}
	},
}

func registerDriveCommands() *cobra.Command {
	driveCmd := &cobra.Command{
		Use:   "drive",
		Short: "drive managment",
	}

	driveCmd.AddCommand(driveMountCmd)
	driveCmd.AddCommand(driveUmountCmd)
	return driveCmd
}
