package launcher

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

var log = logger.NamedLogger("launcher")

// RegisterCmds ...
func RegisterCmds(rootCmd *cobra.Command) {
	startupCmd := &cobra.Command{
		Use:    "launch:startup",
		Hidden: true,
		Short:  "Init commands on sway startup",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			time.Sleep(time.Second * 2)
			for _, cmd := range ctx.EnvironmentConfig.Init {
				_, initJobErr := exec.Command().WithCwd(cmd.Cwd).Start(cmd.Args[0], cmd.Args[1:]...)
				if initJobErr != nil {
					log.Error(initJobErr.Error())
				}
			}
		},
	}

	launcherCmdParams := launchJobParams{}
	launcherCmd := &cobra.Command{
		Use:   "launch",
		Short: "execute cmds based on config file",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			log.Infof("Launching job %s", launcherCmdParams.jobID)
			launcher, launcherErr := createLauncher(ctx)
			if launcherErr != nil {
				log.Errorf("Launcher cration failed")
			}
			if err := launcher.launchJob(launcherCmdParams); err != nil {
				log.Errorf("Launch of the job failed with error [%v]", err)
			}
		},
	}
	launcherCmd.PersistentFlags().StringVar(
		&launcherCmdParams.jobID, "job", "default",
		"job name",
	)
	launcherCmd.PersistentFlags().BoolVar(
		&launcherCmdParams.restart, "restart", false, "restart everything",
	)

	internalLauncherCmdParams := launchTaskParams{}
	internalLauncherCmd := &cobra.Command{
		Use:    "launch:internal",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			launcher, launcherErr := createLauncher(ctx)
			if launcherErr != nil {
				log.Errorf("Launcher creation failed")
				time.Sleep(time.Second * 10) // preserve terminal window before it's closed
			}
			launcher.launchInternalTask(internalLauncherCmdParams)
		},
	}
	internalLauncherCmd.PersistentFlags().StringVar(
		&internalLauncherCmdParams.taskID, "task", "default",
		"task name",
	)
	internalLauncherCmd.PersistentFlags().StringVar(
		&internalLauncherCmdParams.jobID, "job", "default",
		"job name",
	)
	internalLauncherCmd.PersistentFlags().BoolVar(
		&internalLauncherCmdParams.restart, "restart", false, "restart everything",
	)

	launcherStateCmd := &cobra.Command{
		Use:    "launch:internal:state",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			launcher, launcherErr := createLauncher(ctx)
			if launcherErr != nil {
				log.Errorf("Launcher creation failed")
			}
			if err := launcher.printLauncherState(); err != nil {
				log.Errorf("Launch of the job failed with error [%v]", err)
			}
		},
	}

	rootCmd.AddCommand(launcherCmd)
	rootCmd.AddCommand(startupCmd)
	rootCmd.AddCommand(internalLauncherCmd)
	rootCmd.AddCommand(launcherStateCmd)
}
