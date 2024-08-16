package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/fn"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

func RegisterNixCmds(rootCmd *cobra.Command) {
	nixCmds := &cobra.Command{
		Use:   "nix",
		Short: "nix helper",
	}

	shells := []string{
		"membrane",
		"devops",
		"elixir",
		"rust",
	}

	templates := []string{"rust"}

	nixShell := &cobra.Command{
		Use:   "shell",
		Short: "start one of the global shells",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			dotfilesDir := ctx.FromHome(".dotfiles")
			selectedShell, didSelect := prompt.SelectPrompt(
				"Select shell:",
				shells,
				fn.Identity,
			)
			if didSelect {
				runErr := exec.Command().
					WithStdio().
					Args("nix", "develop", fmt.Sprintf("%s#%s", dotfilesDir, selectedShell)).
					Run()
				if runErr != nil {
					log.Error(runErr)
				}
			}
		},
	}

	nixRebuildConfig := &cobra.Command{
		Use:   "rebuild",
		Short: "rebuild current environment",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			cwd := ctx.FromHome(".dotfiles")
			var rebuildCmd *exec.Cmd
			switch os.Getenv("CURRENT_ENV") {
			case "home":
				rebuildCmd = exec.Command().WithStdio().WithCwd(cwd).WithSudo().
					Args("nixos-rebuild", "switch", "--flake", ".#home")
			case "usbdrive":
				rebuildCmd = exec.Command().WithStdio().WithCwd(cwd).WithSudo().
					Args("nixos-rebuild", "switch", "--flake", ".#usbdrive")
			case "work":
				rebuildCmd = exec.Command().WithStdio().WithCwd(cwd).
					Args("home-manager", "switch", "--flake", ".#work")
			case "macbook":
				rebuildCmd = exec.Command().WithStdio().WithCwd(cwd).
					Args("darwin-rebuild", "switch", "--flake", ".#work-mac")
			default:
				log.Warn("nix rebuild not supported in this environment.")
				return
			}

			runErr := exec.RunAll(
				exec.Command().WithCwd(cwd).WithStdio().Args("git", "add", "-A"),
				rebuildCmd,
			)
			if runErr != nil {
				log.Error(runErr)
			}
		},
	}

	nixTemplate := &cobra.Command{
		Use:   "template",
		Short: "generate from template",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			cwd := ctx.FromHome(".dotfiles")

			selectedTemplate, didSelect := prompt.SelectPrompt(
				"Select template:",
				templates,
				fn.Identity,
			)
			if !didSelect {
				return
			}
			runErr := exec.Command().WithStdio().
				Args("nix", "flake", "init", "-t", fmt.Sprintf("%s#%s", cwd, selectedTemplate)).Run()

			if runErr != nil {
				log.Error(runErr)
			}
		},
	}

	nixCmds.AddCommand(nixShell)
	nixCmds.AddCommand(nixRebuildConfig)
	nixCmds.AddCommand(nixTemplate)

	rootCmd.AddCommand(nixCmds)
}
