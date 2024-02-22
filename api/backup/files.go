package backup

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

func fileSync(src string, dst string) error {
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
		Args(
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
		).
		Run()
}

func backupFiles(rootDir string, mapPaths map[string]string) error {
	dirPath := path.Join(rootDir, "files")
	if err := cmd().Args("mkdir", "-p", dirPath).Run(); err != nil {
		return err
	}

	for srcPath, destinationPath := range mapPaths {
		if file.Exists(srcPath) {
			syncErr := fileSync(
				srcPath,
				path.Join(dirPath, destinationPath),
			)
			if syncErr != nil {
				return syncErr
			}
		}
	}
	return nil
}

func restoreFiles(rootDir string, mapPaths map[string]string) error {
	dirPath := path.Join(rootDir, "files")
	if err := cmd().Args("mkdir", "-p", dirPath).Run(); err != nil {
		return err
	}

	for srcPath, destinationPath := range mapPaths {
		if file.Exists(path.Join(dirPath, destinationPath)) {
			syncErr := fileSync(
				path.Join(dirPath, destinationPath),
				srcPath,
			)
			if syncErr != nil {
				return syncErr
			}
		}
	}
	return nil
}
