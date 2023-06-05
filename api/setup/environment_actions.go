package setup

import (
	"path"

	a "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/language"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
)

type SetupLanguageToolchainActionOpts struct {
	Reinstall bool
}

type SetupLspActionOpts struct {
	Reinstall bool
}

func SetupLanguageToolchainAction(ctx context.Context, opts SetupLanguageToolchainActionOpts) a.Object {
	return a.List{
		a.WithCondition{
			If: a.CommandExists("go"),
			Then: a.List{
				language.GoInstallAction("modd", "github.com/cortesi/modd/cmd/modd@latest", opts.Reinstall),
				language.GoInstallAction("golines", "github.com/segmentio/golines@latest", opts.Reinstall),
				language.GoInstallAction("gofumpt", "mvdan.cc/gofumpt@latest", opts.Reinstall),
				language.GoInstallAction(
					"golangci-lint",
					"github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
					opts.Reinstall,
				),
			},
		},
		a.WithCondition{
			// can't check for cmake-format because lsp server also provides executable with that name
			If:   a.Or(a.Not(a.CommandExists("cmake-lint")), a.Const(opts.Reinstall)),
			Then: a.ShellCommand("pip3", "install", "cmakelang"),
		},
	}
}

func SetupLspAction(ctx context.Context, opts SetupLspActionOpts) a.Object {
	return a.Optional{
		Object: a.List{
			a.WithCondition{
				If: a.CommandExists("go"),
				Then: a.List{
					language.GoInstallAction("gopls", "golang.org/x/tools/gopls@latest", opts.Reinstall),
					language.GoInstallAction(
						"efm-langserver",
						"github.com/mattn/efm-langserver@latest",
						opts.Reinstall,
					),
					language.GoInstallAction(
						"golangci-lint-langserver",
						"github.com/nametake/golangci-lint-langserver@latest",
						opts.Reinstall,
					),
				},
			},
			a.WithCondition{
				If: a.CommandExists("npm"),
				Then: a.List{
					language.NodePackageInstallAction("typescript-language-server", opts.Reinstall),
					language.NodePackageInstallAction("typescript", opts.Reinstall),
					language.NodePackageInstallAction("eslint_d", opts.Reinstall),
					language.NodePackageInstallAction("vscode-langservers-extracted", opts.Reinstall),
					language.NodePackageInstallAction("yaml-language-server", opts.Reinstall),
				},
			},
			a.WithCondition{
				If:   a.Or(a.Not(a.CommandExists("cmake-language-server")), a.Const(opts.Reinstall)),
				Then: a.ShellCommand("pip3", "install", "cmake-language-server"),
			},
			nvim.LuaLspInstallAction(ctx, "6ef1608d857e0179c4db7a14037df84dbef676c8"),
		},
	}
}

func SetupEnvironmentCoreAction(ctx context.Context) a.Object {
	return a.List{
		a.WithCondition{
			If: a.Not(
				a.PathExists(path.Join(ctx.Homedir, ".oh-my-zsh")),
			),
			Then: a.ShellCommand("bash",
				"-c",
				"curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh | bash",
			),
		},
		a.ShellCommand("mkdir", "-p", ctx.FromHome(".config")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/sway"), ctx.FromHome(".config/sway")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/i3"), ctx.FromHome(".config/i3")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/alacritty.yml"), ctx.FromHome(".alacritty.yml")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/zshrc"), ctx.FromHome(".zshrc")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/vimrc"), ctx.FromHome(".vimrc")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/ideavimrc"), ctx.FromHome(".ideavimrc")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/nvim"), ctx.FromHome(".config/nvim")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/gtk-3.0"), ctx.FromHome(".config/gtk-3.0")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/docker"), ctx.FromHome(".docker")),
		a.ShellCommand("mkdir", "-p", ctx.FromHome("Dropbox/notes")),
		a.EnsureSymlink(ctx.FromHome("Dropbox/notes"), ctx.FromHome("notes")),
		a.EnsureSymlink(ctx.FromEnvDir("gitconfig"), ctx.FromHome(".gitconfig")),
		a.EnsureSymlink(ctx.FromEnvDir("gitignore"), ctx.FromHome(".gitignore")),
		a.EnsureSymlink(ctx.FromHome(".dotfiles/configs/direnv"), ctx.FromHome(".config/direnv")),
		a.Scope(func() a.Object {
			if ctx.EnvironmentConfig.CustomSetupAction != nil {
				return ctx.EnvironmentConfig.CustomSetupAction(ctx)
			}
			return a.Nop()
		}),
	}
}
