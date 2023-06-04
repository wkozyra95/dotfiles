package setup

import (
	"os"
	"strings"

	a "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
)

type SetupEnvironmentOptions struct {
	Reinstall bool
}

func SetupEnvironment(ctx context.Context, opts SetupEnvironmentOptions) error {
	cmds := a.List{
		a.List{
			ctx.PkgInstaller.EnsurePackagerAction(ctx.Homedir),
			api.PackageInstallAction([]api.Package{
				ctx.PkgInstaller.ShellTools(),
				ctx.PkgInstaller.DevelopmentTools(),
				ctx.PkgInstaller.Desktop(),
			}),
		},
		SetupLanguageToolchainAction(ctx, SetupLanguageToolchainActionOpts(opts)),
		SetupLspAction(ctx, SetupLspActionOpts(opts)),
		a.WithCondition{
			If: a.Not(a.FuncCond(func() (bool, error) {
				return strings.Contains(os.Getenv("SHELL"), "zsh"), nil
			})),
			Then: a.ShellCommand("sudo", "chsh", "-s", "/usr/bin/zsh"),
		},
		SetupEnvironmentCoreAction(ctx),
		nvim.NvimEnsureLazyNvimInstalled(ctx),
		nvim.NvimInstallAction(ctx, "fdc8e966a9183c08f2afec0817d03b7417a883b3"),
	}
	return a.Run(cmds)
}
