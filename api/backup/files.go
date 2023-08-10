package backup

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

func fileSyncAction(src string, dst string) action.Object {
	return action.SimpleAction{
		Run: func() error {
			fileInfo, statErr := os.Stat(src)
			if statErr != nil {
				return statErr
			}
			if fileInfo.IsDir() {
				src = fmt.Sprintf("%s/", src)
				dst = fmt.Sprintf("%s/", dst)
			}
			return exec.Command().
				WithStdio().
				Run(
					"bash", "-c",
					strings.Join([]string{
						"rsync",
						"--update",
						"--delete",
						"--progress",
						"--recursive",
						"--perms",
						"--filter=':- .gitignore'",
						src, dst,
					}, " "),
				)
		},
		Label: fmt.Sprintf("Syncing files %s -> %s", src, dst),
	}
}

func backupFilesAction(rootDir string, mapPaths map[string]string) action.Object {
	dirPath := path.Join(rootDir, "files")
	rsyncActions := []action.Object{}
	for srcPath, destinationPath := range mapPaths {
		rsyncActions = append(rsyncActions, action.WithCondition{
			If: action.PathExists(srcPath),
			Then: fileSyncAction(
				srcPath,
				path.Join(dirPath, destinationPath),
			),
		})
	}
	return append(
		action.List{action.ShellCommand("mkdir", "-p", dirPath)},
		rsyncActions...,
	)
}

func restoreFilesAction(rootDir string, mapPaths map[string]string) action.Object {
	dirPath := path.Join(rootDir, "files")
	rsyncActions := []action.Object{}
	for srcPath, destinationPath := range mapPaths {
		rsyncActions = append(rsyncActions, action.WithCondition{
			If: action.PathExists(path.Join(dirPath, destinationPath)),
			Then: fileSyncAction(
				path.Join(dirPath, destinationPath),
				srcPath,
			),
		})
	}
	return append(
		action.List{action.ShellCommand("mkdir", "-p", dirPath)},
		rsyncActions...,
	)
}
