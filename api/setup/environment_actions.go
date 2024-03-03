package setup

import (
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/language"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

type SetupLanguageToolchainActionOpts struct {
	Reinstall bool
}

type SetupLspActionOpts struct {
	Reinstall bool
}

func SetupLanguageToolchain(ctx context.Context, reinstall bool) error {
	if exec.CommandExists("go") {
		type goPkg struct {
			exec string
			pkg  string
		}
		pkgs := []goPkg{
			{"modd", "github.com/cortesi/modd/cmd/modd@latest"},
			{"golines", "github.com/segmentio/golines@latest"},
			{"gofumpt", "mvdan.cc/gofumpt@latest"},
			{"golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"},
		}
		for _, pkg := range pkgs {
			if !exec.CommandExists(pkg.exec) || reinstall {
				if err := cmd().Args("go", "install", pkg.pkg).Run(); err != nil {
					return err
				}
			}
		}
	}
	if !exec.CommandExists("cmake-lint") || reinstall {
		if err := cmd().Args("pipx", "install", "cmakelang").Run(); err != nil {
			return err
		}
	}
	return nil
}

func InstallLSP(ctx context.Context, opts SetupLspActionOpts) error {
	if exec.CommandExists("go") {
		type goPkg struct {
			exec string
			pkg  string
		}
		pkgs := []goPkg{
			{"gopls", "golang.org/x/tools/gopls@latest"},
			{"efm-langserver", "github.com/mattn/efm-langserver@latest"},
			{"golangci-lint-langserver", "github.com/nametake/golangci-lint-langserver@latest"},
		}
		for _, pkg := range pkgs {
			if !exec.CommandExists(pkg.exec) || opts.Reinstall {
				if err := cmd().Args("go", "install", pkg.pkg).Run(); err != nil {
					return err
				}
			}
		}
	}

	if exec.CommandExists("npm") {
		pkgs := []string{
			"typescript-language-server",
			"typescript",
			"eslint_d",
			"vscode-langservers-extracted",
			"yaml-language-server",
		}
		for _, pkg := range pkgs {
			if err := language.EnsureNodePackageInstalled(pkg, opts.Reinstall); err != nil {
				return err
			}
		}
	}

	if !exec.CommandExists("cmake-language-server") {
		err := cmd().Args("pipx", "install", "cmake-language-server").Run()
		if err != nil {
			return err
		}
	}

	if err := nvim.InstallLuaLSP(ctx, "b96ab075f43e04d5bb42566df4f7c172b35a3df8"); err != nil {
		return err
	}

	return nil
}

func SetupConfigFiles(ctx context.Context) error {
	if err := cmd().Args("mkdir", "-p", ctx.FromHome(".config")).Run(); err != nil {
		return err
	}
	symlinks := []struct {
		src string
		dst string
	}{
		{ctx.FromHome(".dotfiles/configs/sway"), ctx.FromHome(".config/sway")},
		{ctx.FromHome(".dotfiles/configs/i3"), ctx.FromHome(".config/i3")},
		{ctx.FromHome(".dotfiles/configs/alacritty.yml"), ctx.FromHome(".alacritty.yml")},
		{ctx.FromHome(".dotfiles/configs/zshrc"), ctx.FromHome(".zshrc")},
		{ctx.FromHome(".dotfiles/configs/vimrc"), ctx.FromHome(".vimrc")},
		{ctx.FromHome(".dotfiles/configs/ideavimrc"), ctx.FromHome(".ideavimrc")},
		{ctx.FromHome(".dotfiles/configs/nvim"), ctx.FromHome(".config/nvim")},
		{ctx.FromHome(".dotfiles/configs/docker"), ctx.FromHome(".docker")},
	}

	for _, symlink := range symlinks {
		if err := file.EnsureSymlink(symlink.src, symlink.dst); err != nil {
			return err
		}
	}
	return nil
}
