package setup

import (
	"os"
	"path"
	"strings"

	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/platform"
	"github.com/wkozyra95/dotfiles/api/setup/nvim"
	"github.com/wkozyra95/dotfiles/utils/file"
)

type SetupEnvironmentOptions struct {
	Reinstall bool
	DryRun    bool
}

func SetupEnvironment(ctx context.Context, opts SetupEnvironmentOptions) error {
	if os.Getenv("FORCE_MANUAL_SETUP") == "" {
		return setupEnvironmentWithNix(ctx, opts)
	} else {
		return setupEnvironmentManually(ctx, opts)
	}
}

func setupEnvironmentWithNix(ctx context.Context, opts SetupEnvironmentOptions) error {
	log.Warn("Env configured with nix, run \"mycli nix rebuild\" or re-run this command with FORCE_MANUAL_SETUP env.")
	return nil
}

func setupEnvironmentManually(ctx context.Context, opts SetupEnvironmentOptions) error {
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
		pkgInstaller.Desktop(),
	})
	if packageInstallErr != nil {
		return packageInstallErr
	}

	if !strings.Contains(os.Getenv("SHELL"), "zsh") {
		if err := sudo().Args("chsh", "-s", "/usr/bin/zsh").Run(); err != nil {
			return err
		}
	}

	if !file.Exists(path.Join(ctx.Homedir, ".oh-my-zsh")) {
		cmd := cmd().Args(
			"bash",
			"-c",
			"curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh | bash",
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if !file.Exists(ctx.FromHome(".dotfiles-private")) {
		cmd := cmd().Args(
			"git",
			"clone",
			"git@github.com:wkozyra95/dotfiles-private.git",
			ctx.FromHome(".dotfiles-private"),
		)
		if err := cmd.Run(); err != nil {
			return err
		}
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

	if err := file.EnsureSymlink(
		ctx.FromHome(".dotfiles-private/nvim/spell"),
		ctx.FromHome(".dotfiles/configs/nvim/spell"),
	); err != nil {
		return err
	}

	if err := file.EnsureSymlink(ctx.FromHome(".dotfiles-private/notes"), ctx.FromHome("notes")); err != nil {
		return err
	}

	if ctx.EnvironmentConfig.CustomSetupAction != nil {
		if err := ctx.EnvironmentConfig.CustomSetupAction(ctx); err != nil {
			return err
		}
	}
	return nil
}
