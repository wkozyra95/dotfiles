package setup

import (
	"os"
	"strings"

	. "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
)

type SetupEnvironmentOptions struct {
	Reinstall bool
	DryRun    bool
}

func SetupEnvironment(ctx context.Context, opts SetupEnvironmentOptions) error {
	cmds := List{
		List{
			ctx.PkgInstaller.EnsurePackagerAction(ctx.Homedir),
			api.PackageInstallAction([]api.Package{
				ctx.PkgInstaller.ShellTools(),
				ctx.PkgInstaller.DevelopmentTools(),
				ctx.PkgInstaller.Desktop(),
			}),
		},
		SetupLanguageToolchainAction(ctx, SetupLanguageToolchainActionOpts{Reinstall: opts.Reinstall}),
		SetupLspAction(ctx, SetupLspActionOpts{Reinstall: opts.Reinstall}),
		WithCondition{
			If: FuncCond("current shell is not zsh", func() (bool, error) {
				return !strings.Contains(os.Getenv("SHELL"), "zsh"), nil
			}),
			Then: ShellCommand("sudo", "chsh", "-s", "/usr/bin/zsh"),
		},
		WithCondition{
			If: Not(PathExists(ctx.FromHome(".dotfiles-private"))),
			Then: ShellCommand(
				"git",
				"clone",
				"git@github.com:wkozyra95/dotfiles-private.git",
				ctx.FromHome(".dotfiles-private"),
			),
		},
		EnsureSymlink(ctx.FromHome(".dotfiles-private/nvim/spell"), ctx.FromHome(".dotfiles/configs/nvim/spell")),
		EnsureSymlink(ctx.FromHome(".dotfiles-private/notes"), ctx.FromHome("notes")),
		SetupEnvironmentCoreAction(ctx),
		nvim.NvimEnsureLazyNvimInstalled(ctx),
		nvim.NvimInstallAction(ctx, "30d311ebcf9433f84bd4d98f9e049b36c9d352ac"),
	}
	if opts.DryRun {
		PrintActionTree(cmds)
		return nil
	} else {
		return RunActions(cmds)
	}
}
