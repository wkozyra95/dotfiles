package installer

import (
	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
	"github.com/wkozyra95/dotfiles/utils/http"
)

type DownloadInstallOptions struct {
	Path        string
	ArchivePath string
	Url         string
	Reinstall   bool
}

func DownloadZipInstallAction(
	ctx context.Context,
	options DownloadInstallOptions,
	installAction action.Object,
) action.Object {
	return action.List{
		action.WithCondition{
			If: action.Or(action.Not(action.PathExists(options.Path)), action.Const(options.Reinstall)),
			Then: action.List{
				action.ShellCommand("rm", "-rf", options.Path, options.ArchivePath),
				action.DownloadFile(options.Url, options.ArchivePath),
				action.ShellCommand("unzip", "-d", options.Path, options.ArchivePath),
			},
		},
	}
}

func InstallFromZip(
	ctx context.Context,
	options DownloadInstallOptions,
) error {
	if file.Exists(options.Path) && !options.Reinstall {
		return nil
	}
	if err := exec.Command().WithStdio().Args(
		"rm", "-rf", options.Path, options.ArchivePath).Run(); err != nil {
		return err
	}
	if err := http.DownloadFile(options.Url, options.ArchivePath); err != nil {
		return err
	}
	if err := exec.Command().WithStdio().Args("unzip", "-d", options.Path, options.ArchivePath).Run(); err != nil {
		return err
	}
	return nil
}
