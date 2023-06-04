package command

import (
	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/docker"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

func RegisterDockerCmds(rootCmd *cobra.Command) {
	dockerCmds := &cobra.Command{
		Use:   "docker",
		Short: "docker helper",
	}

	dockerRunOptions := docker.DockerRunOptions{}
	dockerRun := &cobra.Command{
		Use:   "run",
		Short: "start container",
		Run: func(cmd *cobra.Command, args []string) {
			cliCtx := context.CreateContext()
			ctx := docker.NewDockerContext()
			selected, didSelect := prompt.SelectPrompt(
				"Select docker image",
				cliCtx.EnvironmentConfig.DockerEnvsSpec,
				func(spec env.DockerEnvSpec) string { return spec.Name },
			)
			if !didSelect {
				log.Error("Nothing was selected")
				return
			}
			if err := docker.Run(ctx, selected, dockerRunOptions); err != nil {
				log.Error(err.Error())
			}
		},
	}

	dockerCmds.AddCommand(dockerRun)
	dockerCmds.PersistentFlags().BoolVar(&dockerRunOptions.Rebuild, "rebuild", false, "rebuild image and containers")

	rootCmd.AddCommand(dockerCmds)
}
