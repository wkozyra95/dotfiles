package installer

import (
	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
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
