package setup

import (
	"os"

	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/platform"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
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

	if exec.CommandExists("go") {
		if err := cmd().Args("wget", "-P", "/tmp", "https://go.dev/dl/go1.20.2.linux-amd64.tar.gz").Run(); err != nil {
			return err
		}
		unpackCmd := cmd().Args("tar", "-C", "/usr/local", "-xzf", "/tmp/go1.20.2.linux-amd64.tar.gz")
		if ctx.Username == "root" {
			unpackCmd = unpackCmd.WithSudo()
		}
		if err := unpackCmd.Run(); err != nil {
			return err
		}
		os.Setenv("PATH", "/usr/local/go/bin:"+os.Getenv("PATH"))
	}

	if err := SetupLanguageToolchain(ctx, opts.Reinstall); err != nil {
		return err
	}

	if err := InstallLSP(ctx, SetupLspActionOpts{Reinstall: opts.Reinstall}); err != nil {
		return err
	}

	if err := nvim.EnsureLazyNvimInstalled(ctx); err != nil {
		return err
	}

	if err := nvim.InstallNvimFromSource(ctx, "c0cb1e8e9437b738c8d3232ec4594113d2221bb2"); err != nil {
		return err
	}

	if err := SetupConfigFiles(ctx); err != nil {
		return err
	}

	if !file.Exists(ctx.FromHome(".dotfiles")) {
		err := exec.RunAll(
			cmd().Args(
				"git",
				"clone",
				"https://github.com/wkozyra95/dotfiles.git",
				ctx.FromHome(".dotfiles"),
			),
			cmd().WithCwd(ctx.FromHome(".dotfiles")).Args("make"),
		)
		if err != nil {
			return err
		}
	}

	if !file.Exists(ctx.FromHome(".fzf")) {
		cmd := cmd().Args(
			"git",
			"clone",
			"--depth", "1",
			"https://github.com/junegunn/fzf.git",
			ctx.FromHome(".fzf"),
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if !file.Exists(ctx.FromHome(".oh-my-zsh")) {
		cmd := cmd().Args(
			"bash",
			"-c",
			"curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh | bash",
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
