package setup

import (
	"os"
	"path"

	a "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/platform"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

func SetupUbuntuInDocker(ctx context.Context, opts SetupEnvironmentOptions) error {
	pkgInstaller, pkgInstallerErr := platform.GetPackageManager(ctx)
	if pkgInstallerErr != nil {
		return pkgInstallerErr
	}
	if err := pkgInstaller.EnsurePackagerInstalled(ctx.Homedir); err != nil {
		return err
	}
	packageInstallErr := api.InstallPackages([]api.Package{
		pkgInstaller.ShellTools(),
		pkgInstaller.DevelopmentTools(),
	})

	if packageInstallErr != nil {
		return packageInstallErr
	}
	cmds := a.List{
		a.WithCondition{
			If: a.Not(a.CommandExists("go")),
			Then: a.List{
				a.ShellCommand("wget", "-P", "/tmp", "https://go.dev/dl/go1.20.2.linux-amd64.tar.gz"),
				a.WithCondition{
					If:   a.Const(ctx.Username == "root"),
					Then: a.ShellCommand("tar", "-C", "/usr/local", "-xzf", "/tmp/go1.20.2.linux-amd64.tar.gz"),
					Else: a.ShellCommand("sudo", "tar", "-C", "/usr/local", "-xzf", "/tmp/go1.20.2.linux-amd64.tar.gz"),
				},
				a.Func("Add /usr/local/go/bin to PATH", func() error {
					os.Setenv("PATH", "/usr/local/go/bin:"+os.Getenv("PATH"))
					return nil
				}),
			},
		},
		SetupLanguageToolchainAction(ctx, SetupLanguageToolchainActionOpts{Reinstall: opts.Reinstall}),
		SetupLspAction(ctx, SetupLspActionOpts{Reinstall: opts.Reinstall}),
		a.WithCondition{
			If: a.Not(a.PathExists(ctx.FromHome(".dotfiles"))),
			Then: a.List{
				a.ShellCommand(
					"git",
					"clone",
					"https://github.com/wkozyra95/dotfiles.git",
					ctx.FromHome(".dotfiles"),
				),
				a.Execute(exec.Command().WithCwd(ctx.FromHome(".dotfiles")), "make"),
			},
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
		a.WithCondition{
			If: a.Not(
				a.PathExists(path.Join(ctx.Homedir, ".oh-my-zsh")),
			),
			Then: a.ShellCommand("bash",
				"-c",
				"curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh | bash",
			),
		},
		SetupEnvironmentCoreAction(ctx),
		nvim.NvimEnsureLazyNvimInstalled(ctx),
		nvim.NvimInstallAction(ctx, "fdc8e966a9183c08f2afec0817d03b7417a883b3"),
	}
	return a.RunActions(cmds, false)
}
