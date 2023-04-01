package backup

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

type fileSyncActionArgs struct {
	src string
	dst string
}

var fileSyncAction = action.SimpleActionBuilder[fileSyncActionArgs]{
	CreateRun: func(args fileSyncActionArgs) func() error {
		return func() error {
			fileInfo, statErr := os.Stat(args.src)
			if statErr != nil {
				return statErr
			}
			src := args.src
			dst := args.dst
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
		}
	},
	String: func(args fileSyncActionArgs) string {
		return fmt.Sprintf("Syncing files %s -> %s", args.src, args.dst)
	},
}.Init()

func backupFilesAction(rootDir string, mapPaths map[string]string) action.Object {
	dirPath := path.Join(rootDir, "files")
	rsyncActions := []action.Object{}
	for srcPath, destinationPath := range mapPaths {
		rsyncActions = append(rsyncActions, action.WithCondition{
			If: action.PathExists(srcPath),
			Then: fileSyncAction(fileSyncActionArgs{
				src: srcPath,
				dst: path.Join(dirPath, destinationPath),
			}),
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
			Then: fileSyncAction(fileSyncActionArgs{
				src: path.Join(dirPath, destinationPath),
				dst: srcPath,
			}),
		})
	}
	return append(
		action.List{action.ShellCommand("mkdir", "-p", dirPath)},
		rsyncActions...,
	)
}
