package nvim

import (
	"fmt"
	"path"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/setup/git"
	"github.com/wkozyra95/dotfiles/api/setup/installer"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

func NvimInstallAction(ctx context.Context, commitHash string) action.Object {
	withCwd := func(path string) *exec.Cmd {
		return exec.Command().WithStdio().WithCwd(path)
	}
	cloneDir := ctx.FromHome(".cache/nvim_source")
	return git.RepoInstallAction(ctx, git.RepoInstallOptions{
		Path:       cloneDir,
		Name:       "nvim",
		CommitHash: commitHash,
		RepoUrl:    "https://github.com/neovim/neovim.git",
	}, action.List{
		action.Execute(
			withCwd(cloneDir),
			"make",
			"-j12",
			"CMAKE_BUILD_TYPE=RelWithDebInfo",
			fmt.Sprintf("CMAKE_INSTALL_PREFIX=%s", ctx.FromHome(".local")),
		),
		action.Execute(withCwd(cloneDir), "make", "install"),
	})
}

func ElixirLspInstallAction(ctx context.Context, reinstall bool) action.Object {
	return installer.DownloadZipInstallAction(
		ctx,
		installer.DownloadInstallOptions{
			Path:        ctx.FromHome(".cache/nvim/myconfig/elixirls"),
			ArchivePath: ctx.FromHome(".cache/nvim/myconfig/elixirls.zip"),
			Url:         "https://github.com/elixir-lsp/elixir-ls/releases/download/v0.15.0/elixir-ls-v0.15.0.zip",
			Reinstall:   reinstall,
		},
		action.ShellCommand("chmod", "+x", ctx.FromHome(".cache/nvim/myconfig/elixirls/language_server.sh")),
	)
}

func LuaLspInstallAction(ctx context.Context, commitHash string) action.Object {
	withCwd := func(path string) *exec.Cmd {
		return exec.Command().WithStdio().WithCwd(path)
	}
	cloneDir := ctx.FromHome(".cache/nvim/myconfig/lua_lsp")
	return git.RepoInstallAction(ctx, git.RepoInstallOptions{
		Path:       cloneDir,
		Name:       "lua_lsp",
		CommitHash: commitHash,
		RepoUrl:    "https://github.com/LuaLS/lua-language-server",
	}, action.List{
		action.Execute(
			withCwd(cloneDir),
			"git", "submodule", "update", "--init", "--recursive",
		),
		action.Execute(
			withCwd(path.Join(cloneDir, "/3rd/luamake")).WithEnv("ZDOTDIR=/tmp"),
			"bash",
			"./compile/install.sh",
		),
		action.Execute(
			withCwd(cloneDir),
			"./3rd/luamake/luamake", "rebuild",
		),
		action.Execute(withCwd(path.Join(cloneDir, "3rd/EmmyLuaCodeStyle")), "mkdir", "build"),
		action.Execute(withCwd(path.Join(cloneDir, "3rd/EmmyLuaCodeStyle/build")), "cmake", ".."),
		action.Execute(withCwd(path.Join(cloneDir, "3rd/EmmyLuaCodeStyle/build")), "cmake", "--build", "."),
	})
}

func NvimEnsureLazyNvimInstalled(ctx context.Context) action.Object {
	return action.WithCondition{
		If: action.Not(
			action.PathExists(path.Join(ctx.Homedir, ".local/share/nvim/lazy/lazy.nvim")),
		),
		Then: action.ShellCommand(
			"bash",
			"-c",
			"git clone --filter=blob:none https://github.com/folke/lazy.nvim.git --branch=stable ~/.local/share/nvim/lazy/lazy.nvim",
		),
	}
}
