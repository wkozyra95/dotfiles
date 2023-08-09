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
	reinstallCond := a.LabeledConst("reinstall", opts.Reinstall)
	return a.List{
		a.WithCondition{
			If: a.CommandExists("go"),
			Then: a.List{
				a.WithCondition{
					If:   a.Or(a.Not(a.CommandExists("modd")), reinstallCond),
					Then: a.ShellCommand("go", "install", "github.com/cortesi/modd/cmd/modd@latest"),
				},
				a.WithCondition{
					If:   a.Or(a.Not(a.CommandExists("golines")), reinstallCond),
					Then: a.ShellCommand("go", "install", "github.com/segmentio/golines@latest"),
				},
				a.WithCondition{
					If:   a.Or(a.Not(a.CommandExists("gofumpt")), reinstallCond),
					Then: a.ShellCommand("go", "install", "mvdan.cc/gofumpt@latest"),
				},
				a.WithCondition{
					If:   a.Or(a.Not(a.CommandExists("golangci-lint")), reinstallCond),
					Then: a.ShellCommand("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"),
				},
			},
		},
		a.WithCondition{
			// can't check for cmake-format because lsp server also provides executable with that name
			If:   a.Or(a.Not(a.CommandExists("cmake-lint")), a.LabeledConst("reinstall", opts.Reinstall)),
			Then: a.ShellCommand("pipx", "install", "cmakelang"),
		},
	}
}

func SetupLspAction(ctx context.Context, opts SetupLspActionOpts) a.Object {
	reinstallCond := a.LabeledConst("reinstall", opts.Reinstall)
	return a.Optional(
		"(optional) Setup LSPs",
		a.List{
			a.WithCondition{
				If: a.CommandExists("go"),
				Then: a.List{
					a.WithCondition{
						If:   a.Or(a.Not(a.CommandExists("gopls")), reinstallCond),
						Then: a.ShellCommand("go", "install", "golang.org/x/tools/gopls@latest"),
					},
					a.WithCondition{
						If:   a.Or(a.Not(a.CommandExists("efm-langserver")), reinstallCond),
						Then: a.ShellCommand("go", "install", "github.com/mattn/efm-langserver@latest"),
					},
					a.WithCondition{
						If:   a.Or(a.Not(a.CommandExists("golangci-lint-langserver")), reinstallCond),
						Then: a.ShellCommand("go", "install", "github.com/nametake/golangci-lint-langserver@latest"),
					},
				},
			},
			a.WithCondition{
				If: a.CommandExists("npm"),
				Then: a.List{
					language.NodePackageInstallAction("typescript-language-server", reinstallCond),
					language.NodePackageInstallAction("typescript", reinstallCond),
					language.NodePackageInstallAction("eslint_d", reinstallCond),
					language.NodePackageInstallAction("vscode-langservers-extracted", reinstallCond),
					language.NodePackageInstallAction("yaml-language-server", reinstallCond),
				},
			},
			a.WithCondition{
				If:   a.Or(a.Not(a.CommandExists("cmake-language-server")), reinstallCond),
				Then: a.ShellCommand("pipx", "install", "cmake-language-server"),
			},
			nvim.LuaLspInstallAction(ctx, "b96ab075f43e04d5bb42566df4f7c172b35a3df8"),
		},
	)
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
		a.Scope("Run custom environment hooks", func() a.Object {
			if ctx.EnvironmentConfig.CustomSetupAction != nil {
				return ctx.EnvironmentConfig.CustomSetupAction(ctx)
			}
			return a.Nop()
		}),
	}
}
