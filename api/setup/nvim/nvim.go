package nvim

import (
	"fmt"
	"path"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/setup/git"
	"github.com/wkozyra95/dotfiles/api/setup/installer"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

func cmd() *exec.Cmd {
	return exec.Command().WithStdio()
}

func InstallNvimFromSource(ctx context.Context, commitHash string) error {
	cloneDir := ctx.FromHome(".cache/nvim_source")
	return git.InstallFromRepo(ctx, git.RepoInstallOptions{
		Path:       cloneDir,
		Name:       "nvim",
		CommitHash: commitHash,
		RepoUrl:    "https://github.com/neovim/neovim.git",
	}, func(ctx context.Context) error {
		return exec.RunAll(
			cmd().WithCwd(cloneDir).Args(
				"make",
				"-j12",
				"CMAKE_BUILD_TYPE=RelWithDebInfo",
				fmt.Sprintf("CMAKE_INSTALL_PREFIX=%s", ctx.FromHome(".local")),
			),
			cmd().WithCwd(cloneDir).Args("make", "install"),
		)
	})
}

func InstallElixirLSP(ctx context.Context, reinstall bool) error {
	installErr := installer.InstallFromZip(
		ctx,
		installer.DownloadInstallOptions{
			Path:        ctx.FromHome(".cache/nvim/myconfig/elixirls"),
			ArchivePath: ctx.FromHome(".cache/nvim/myconfig/elixirls.zip"),
			Url:         "https://github.com/elixir-lsp/elixir-ls/releases/download/v0.15.0/elixir-ls-v0.15.0.zip",
			Reinstall:   reinstall,
		},
	)
	if installErr != nil {
		return installErr
	}
	entrypointPath := ctx.FromHome(".cache/nvim/myconfig/elixirls/language_server.sh")
	return cmd().Args("chmod", "+x", entrypointPath).Run()
}

func InstallLuaLSP(ctx context.Context, commitHash string) error {
	cloneDir := ctx.FromHome(".cache/nvim/myconfig/lua_lsp")
	return git.InstallFromRepo(ctx, git.RepoInstallOptions{
		Path:       cloneDir,
		Name:       "lua_lsp",
		CommitHash: commitHash,
		RepoUrl:    "https://github.com/LuaLS/lua-language-server",
	}, func(ctx context.Context) error {
		return exec.RunAll(
			cmd().WithCwd(cloneDir).Args(
				"git", "submodule", "update", "--init", "--recursive",
			),
			cmd().
				WithCwd(path.Join(cloneDir, "/3rd/luamake")).
				WithEnv("ZDOTDIR=/tmp").
				Args("bash", "./compile/install.sh"),
			cmd().WithCwd(cloneDir).Args("./3rd/luamake/luamake", "rebuild"),
			cmd().WithCwd(path.Join(cloneDir, "3rd/EmmyLuaCodeStyle")).Args("mkdir", "build"),
			cmd().WithCwd(path.Join(cloneDir, "3rd/EmmyLuaCodeStyle/build")).Args("cmake", ".."),
			cmd().WithCwd(path.Join(cloneDir, "3rd/EmmyLuaCodeStyle/build")).Args("cmake", "--build", "."),
		)
	})
}

func EnsureLazyNvimInstalled(ctx context.Context) error {
	if file.Exists(path.Join(ctx.Homedir, ".local/share/nvim/lazy/lazy.nvim")) {
		return nil
	}
	cloneCmd := "git clone --filter=blob:none https://github.com/folke/lazy.nvim.git --branch=stable ~/.local/share/nvim/lazy/lazy.nvim"
	return cmd().Args("bash", "-c", cloneCmd).Run()
}
