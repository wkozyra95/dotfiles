package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

var log = NamedLogger("file")

func LogDirectoryListing(destination string, maxEntries int) {
	file, fileErr := os.Open(destination)
	if fileErr != nil {
		log.Warnf("Directory %s doesn't exist", destination)
		return
	}

	fileInfo, fileInfoErr := file.Stat()
	if fileInfoErr != nil {
		log.Warnf("Directory %s doesn't exist", destination)
		return
	}

	if !fileInfo.IsDir() {
		log.Warnf("%s is not a directory", destination)
		return
	}

	files, readdirErr := file.Readdir(maxEntries)
	if readdirErr != nil && readdirErr != io.EOF {
		log.Warnf("Unable to read content of %s directory [%v]", destination, readdirErr)
		return
	}
	var fileListing strings.Builder
	fileListing.WriteString(fmt.Sprintf("Content of %s directory: \n", destination))
	for _, file := range files {
		fileListing.WriteString(
			fmt.Sprintf("\t%s %6d KB %s\n",
				file.Mode(), file.Size()/1024, file.Name(),
			),
		)
		if file.IsDir() {
			file, fileErr := os.Open(path.Join(destination, file.Name()))
			if fileErr != nil {
				fileListing.WriteString("\t\t Unable to access content\n")
				continue
			}
			files, readdirErr := file.Readdir(10)
			if readdirErr != nil {
				fileListing.WriteString("\t\t Unable to access content\n")
				continue
			}
			for _, file := range files {
				fileListing.WriteString(
					fmt.Sprintf("\t\t%s %6d KB %s\n",
						file.Mode(), file.Size()/1024, file.Name(),
					),
				)
			}
			if len(files) == maxEntries {
				fileListing.WriteString("\t\t...\n")
			}
		}
	}
	if len(files) == maxEntries {
		fileListing.WriteString("\t...\n")
	}
	log.Info(fileListing.String())
}
