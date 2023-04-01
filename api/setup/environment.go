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
		SetupEnvirionmentCoreAction(ctx),
		nvim.NvimEnsureLazyNvimInstalled(ctx),
		nvim.NvimInstallAction(ctx, "65046c830e14f8988d9c3b477187f6b871e45af2"),
	}
	return a.Run(cmds)
}
