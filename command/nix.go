package command

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/fn"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

func rebuildCurrentEnv() {
	ctx := context.CreateContext()
	cwd := ctx.FromHome(".dotfiles")
	var rebuildCmd *exec.Cmd
	switch os.Getenv("CURRENT_ENV") {
	case "home":
		rebuildCmd = exec.Command().WithStdio().WithCwd(cwd).WithSudo().
			Args("nixos-rebuild", "switch", "--flake", ".#home")
	case "home-nas":
		rebuildCmd = exec.Command().WithStdio().WithCwd(cwd).WithSudo().
			Args("nixos-rebuild", "switch", "--flake", ".#home-nas")
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
}

func RegisterNixCmds(rootCmd *cobra.Command) {
	nixCmds := &cobra.Command{
		Use:   "nix",
		Short: "nix helper",
	}

	shells := []string{
		"smelter",
		"membrane",
		"devops",
		"elixir",
		"rust",
		"nodejs",
	}

	templates := []string{"rust", "nodejs"}

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
					// nix develop sets PATH dev-shell-first, but it doesn't carry
					// the nix profile guard, so any shell started inside (e.g. zsh)
					// re-sources /etc/zsh/zshrc -> nix-daemon.sh and re-prepends
					// ~/.nix-profile/bin, shadowing the dev shell. Pre-set the guard
					// so that re-prepend is skipped and dev-shell tools win.
					WithEnv("__ETC_PROFILE_NIX_SOURCED=1").
					// Launch straight into zsh instead of the default bash. The
					// guard above keeps the dev shell ahead of ~/.nix-profile here.
					Args("nix", "develop", fmt.Sprintf("%s#%s", dotfilesDir, selectedShell), "--command", "zsh").
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
			rebuildCurrentEnv()
		},
	}

	nixUpdate := &cobra.Command{
		Use:   "update",
		Short: "update flake inputs of the dotfiles",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			updateCmd := exec.Command().WithStdio().WithCwd(ctx.FromHome(".dotfiles"))

			// Anonymous requests to the GitHub API are limited to 60 per hour.
			var token, stderr bytes.Buffer
			tokenErr := exec.Command().WithBufout(&token, &stderr).Args("gh", "auth", "token").Run()
			if tokenErr != nil {
				log.Warn("Unable to read GitHub token from gh, running unauthenticated.")
			} else {
				// Keep any existing NIX_CONFIG, setting it here would replace it.
				updateCmd.WithEnv(fmt.Sprintf(
					"NIX_CONFIG=%s\naccess-tokens = github.com=%s",
					os.Getenv("NIX_CONFIG"),
					strings.TrimSpace(token.String()),
				))
			}

			if runErr := updateCmd.Args("nix", "flake", "update").Run(); runErr != nil {
				log.Error(runErr)
				return
			}
			if prompt.ConfirmPrompt("Do you want to rebuild now?") {
				rebuildCurrentEnv()
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
	nixCmds.AddCommand(nixUpdate)
	nixCmds.AddCommand(nixTemplate)

	rootCmd.AddCommand(nixCmds)
}
