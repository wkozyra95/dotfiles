package setup

import (
	"os"
	"path"
	"strings"

	. "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/platform"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
)

type SetupEnvironmentOptions struct {
	Reinstall bool
	DryRun    bool
}

func SetupEnvironment(ctx context.Context, opts SetupEnvironmentOptions) error {
	if os.Getenv("FORCE_MANUAL_SETUP") == "" {
		return setupEnvironmentWithNix(ctx, opts)
	} else {
		return setupEnvironmentManually(ctx, opts)
	}
}

func setupEnvironmentWithNix(ctx context.Context, opts SetupEnvironmentOptions) error {
	cmds := List{
		WithCondition{
			If: Not(PathExists(ctx.FromHome(".dotfiles-private"))),
			Then: ShellCommand(
				"git",
				"clone",
				"git@github.com:wkozyra95/dotfiles-private.git",
				ctx.FromHome(".dotfiles-private"),
			),
		},
		nvim.NvimEnsureLazyNvimInstalled(ctx),
		Scope("Run custom environment hooks", func() Object {
			if ctx.EnvironmentConfig.CustomSetupAction != nil {
				return ctx.EnvironmentConfig.CustomSetupAction(ctx)
			}
			return Nop()
		}),
	}
	if opts.DryRun {
		PrintActionTree(cmds)
		return nil
	} else {
		return RunActions(cmds, true)
	}
}

func setupEnvironmentManually(ctx context.Context, opts SetupEnvironmentOptions) error {
	pkgInstaller, pkgInstallerErr := platform.GetPackageManager(ctx)
	if pkgInstallerErr != nil {
		return pkgInstallerErr
	}
	cmds := List{
		List{
			pkgInstaller.EnsurePackagerAction(ctx.Homedir),
			api.PackageInstallAction([]api.Package{
				pkgInstaller.ShellTools(),
				pkgInstaller.DevelopmentTools(),
				pkgInstaller.Desktop(),
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
			If: Not(
				PathExists(path.Join(ctx.Homedir, ".oh-my-zsh")),
			),
			Then: ShellCommand("bash",
				"-c",
				"curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh | bash",
			),
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
		nvim.NvimInstallAction(ctx, "c0cb1e8e9437b738c8d3232ec4594113d2221bb2"),
	}
	if opts.DryRun {
		PrintActionTree(cmds)
		return nil
	} else {
		return RunActions(cmds, true)
	}
}
