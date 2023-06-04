package setup

import (
	"os"

	a "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
)

func SetupUbuntuInDocker(ctx context.Context, opts SetupEnvironmentOptions) error {
	cmds := a.List{
		a.List{
			ctx.PkgInstaller.EnsurePackagerAction(ctx.Homedir),
			api.PackageInstallAction([]api.Package{
				ctx.PkgInstaller.ShellTools(),
				ctx.PkgInstaller.DevelopmentTools(),
			}),
		},
		a.WithCondition{
			If: a.Not(a.CommandExists("go")),
			Then: a.List{
				a.ShellCommand("wget", "-P", "/tmp", "https://go.dev/dl/go1.20.2.linux-amd64.tar.gz"),
				a.WithCondition{
					If:   a.Const(ctx.Username == "root"),
					Then: a.ShellCommand("tar", "-C", "/usr/local", "-xzf", "/tmp/go1.20.2.linux-amd64.tar.gz"),
					Else: a.ShellCommand("sudo", "tar", "-C", "/usr/local", "-xzf", "/tmp/go1.20.2.linux-amd64.tar.gz"),
				},
				a.Func(func() error {
					os.Setenv("PATH", "/usr/local/go/bin:"+os.Getenv("PATH"))
					return nil
				}),
			},
		},
		SetupLanguageToolchainAction(ctx, SetupLanguageToolchainActionOpts(opts)),
		SetupLspAction(ctx, SetupLspActionOpts(opts)),
		a.WithCondition{
			If: a.Not(a.PathExists(ctx.FromHome(".dotfiles"))),
			Then: a.ShellCommand(
				"git",
				"clone",
				"https://github.com/wkozyra95/dotfiles.git",
				ctx.FromHome(".dotfiles"),
			),
		},
		a.WithCondition{
			If: a.Not(a.PathExists(ctx.FromHome(".fzf"))),
			Then: a.ShellCommand(
				"git",
				"clone",
				"--depth", "1",
				"https://github.com/junegunn/fzf.git",
				ctx.FromHome(".fzf"),
			),
		},
		SetupEnvironmentCoreAction(ctx),
		nvim.NvimEnsureLazyNvimInstalled(ctx),
		nvim.NvimInstallAction(ctx, "4e5061dba765df2a74ac4a8182f6e7fe21da125d"),
	}
	return a.Run(cmds)
}
