package file

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

var log = logger.NamedLogger("file")

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Errorf("Accessing file failed with [%s]", err.Error())
	}
	return true
}

func Copy(source string, destination string) error {
	data, readErr := os.ReadFile(source)
	if readErr != nil {
		return readErr
	}
	if err := os.WriteFile(destination, data, 0o644); err != nil {
		return err
	}
	return nil
}

func CopyAsRoot(source string, destination string) error {
	return exec.Command().WithSudo().Args("cp", "-R", source, destination).Run()
}

func EnsureSymlink(source string, destination string) error {
	log.Debugf("EnsureSymlink(%s, %s)", source, destination)
	linkStat, linkStatErr := os.Lstat(destination)
	if linkStatErr != nil {
		if os.IsNotExist(linkStatErr) {
			return os.Symlink(source, destination)
		} else {
			return linkStatErr
		}
	}
	if linkStat.Mode()&fs.ModeSymlink == 0 {
		log.Debug("Replacing files with symlink")
		if err := os.RemoveAll(destination); err != nil {
			return err
		}
		return os.Symlink(source, destination)
	}
	linkDestination, linkDestinationErr := os.Readlink(destination)
	if linkDestinationErr != nil || linkDestination != source {
		log.Debug("Recreating link")
		if err := os.RemoveAll(destination); err != nil {
			return err
		}
		return os.Symlink(source, destination)
	}
	return nil
}

func ensureTextInString(content string, text string, rg *regexp.Regexp) (bool, string) {
	var match []int
	if rg != nil {
		match = rg.FindStringIndex(content)
	}
	if len(match) == 0 {
		if len(content) == 0 {
			return true, fmt.Sprintf("%s\n", text)
		}
		lastCharacter := content[len(content)-1]
		if lastCharacter == '\n' {
			return true, fmt.Sprintf("%s%s\n", content, text)
		} else {
			return true, fmt.Sprintf("%s\n%s\n", content, text)
		}
	} else {
		updatedContent := strings.Join(
			[]string{
				string(content[0:match[0]]),
				text,
				string(content[match[1]:]),
			},
			"",
		)
		return updatedContent != content, updatedContent
	}
}

func EnsureText(path string, text string) error {
	return EnsureTextWithRegexp(path, text, regexp.MustCompile(regexp.QuoteMeta(text)))
}

func EnsureTextWithRegexp(path string, text string, rg *regexp.Regexp) error {
	content := ""
	if Exists(path) {
		byteContent, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		content = string(byteContent)
	}
	shouldUpdate, updatedContent := ensureTextInString(content, text, rg)
	if !shouldUpdate {
		return nil
	}

	file, openErr := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if openErr != nil {
		return openErr
	}
	defer file.Close()
	if _, err := file.WriteString(updatedContent); err != nil {
		return err
	}
	return nil
}
