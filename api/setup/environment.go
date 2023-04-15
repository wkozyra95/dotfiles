package setup

import (
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
		SetupEnvironmentCoreAction(ctx),
		nvim.NvimEnsureLazyNvimInstalled(ctx),
		nvim.NvimInstallAction(ctx, "0a61cb60a633f05e3f05f8d5dae241910a92ee1f"),
	}
	return a.Run(cmds)
}
