package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/fn"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

func ProvisionUsbNixInstaller(ctx context.Context) error {
	target, targetErr := selectPortableInstallMedia()
	if targetErr != nil {
		return targetErr
	}

	dotfilesPath := ctx.FromHome(".dotfiles")

	buildErr := exec.RunAll(
		exec.Command().WithStdio().WithCwd(dotfilesPath).Args("git", "add", "-A"),
		exec.Command().WithStdio().WithCwd(dotfilesPath).Args(
			"nix", "build",
			".#nixosConfigurations.iso-installer.config.system.build.isoImage",
		),
	)

	if buildErr != nil {
		return buildErr
	}

	isoDir := ctx.FromHome(".dotfiles/result")
	files, fileErr := os.ReadDir(isoDir)
	if fileErr != nil {
		return fileErr
	}

	rg := regexp.MustCompile(`.*x86_64-linux\.iso`)
	isoFiles := fn.Filter(files, func(de fs.DirEntry) bool {
		return rg.MatchString(de.Name())
	})

	if len(isoFiles) == 0 {
		return errors.New("no iso files found")
	}
	isoPath := path.Join(isoDir, isoFiles[0].Name())

	if !prompt.ConfirmPrompt(fmt.Sprintf("Do you want to copy iso firl %s to %s device", isoPath, target)) {
		return fmt.Errorf("Aborting ...")
	}
	return exec.Command().WithStdio().Args("dd",
		fmt.Sprintf("if=%s", isoPath),
		fmt.Sprintf("of=%s", target),
		fmt.Sprintf("bs=%dK", 4*1024),
		"status=progress",
		"conv=fsync",
		"oflag=direct",
	).Run()
}
