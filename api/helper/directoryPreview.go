package helper

import (
	"os"
	"path"
)

type DirectoryPreviewOptions struct {
	MaxElements int
}

type FileInfo struct {
	Path         string  `json:"path"`
	IsExecutable bool    `json:"is_executable"`
	IsDirectory  bool    `json:"is_directory"`
	IsHidden     bool    `json:"is_hidden"`
	Name         string  `json:"name"`
	Symlink      *string `json:"symlink"`
}

func GetDirectoryPreview(directory string, opts DirectoryPreviewOptions) ([]FileInfo, error) {
	results := make([]FileInfo, 0, opts.MaxElements)
	dirContent, readDirErr := os.ReadDir(directory)
	if readDirErr != nil {
		return nil, readDirErr
	}
	for _, dirEntry := range dirContent {
		if dirEntry.Name() == ".git" {
			continue
		}
		filePath := path.Join(directory, dirEntry.Name())
		mode := dirEntry.Type()
		symlink := (*string)(nil)
		if mode&os.ModeSymlink != 0 {
			linkDestination, linkDestinationErr := os.Readlink(filePath)
			if linkDestinationErr != nil {
				return nil, linkDestinationErr
			}
			symlink = &linkDestination
		}
		results = append(results, FileInfo{
			Path:         filePath,
			IsExecutable: mode&0o100 != 0,
			IsDirectory:  dirEntry.IsDir(),
			IsHidden:     dirEntry.Name()[0] == '.',
			Name:         dirEntry.Name(),
			Symlink:      symlink,
		})
	}
	orderedResults := make([]FileInfo, 0, len(results))
	for _, maybeDir := range results {
		if maybeDir.IsDirectory {
			orderedResults = append(orderedResults, maybeDir)
		}
	}
	for _, maybeDir := range results {
		if !maybeDir.IsDirectory {
			orderedResults = append(orderedResults, maybeDir)
		}
	}
	return orderedResults, nil
}
