package git

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

type RepoInstallOptions struct {
	Path       string
	RepoUrl    string
	Name       string
	CommitHash string
}

func RepoInstallAction(ctx context.Context, options RepoInstallOptions, installAction action.Object) action.Object {
	withCwd := func(path string) *exec.Cmd {
		return exec.Command().WithCwd(path)
	}
	installPrefix := ctx.FromHome(".local")
	hashFileName := fmt.Sprintf(".%s.hash", options.Name)
	shortPath := fmt.Sprintf("~/.local/%s", hashFileName)
	hashFile := path.Join(installPrefix, hashFileName)
	return action.List{
		action.WithCondition{
			If: action.Not(action.PathExists(options.Path)),
			Then: action.List{
				action.ShellCommand("mkdir", "-p", path.Dir(options.Path)),
				action.ShellCommand("git", "clone", options.RepoUrl, options.Path),
			},
		},
		action.WithCondition{
			If: action.FuncCond(
				fmt.Sprintf("content of %s does not match %s", shortPath, options.CommitHash),
				func() (bool, error) {
					if !file.Exists(hashFile) {
						return true, nil
					}
					file, readErr := os.ReadFile(hashFile)
					if readErr != nil {
						return false, readErr
					}
					return strings.Trim(string(file), "\n ") != options.CommitHash, nil
				}),
			Then: action.List{
				action.ShellCommand("mkdir", "-p", installPrefix),
				action.Execute(withCwd(options.Path), "git", "fetch", "origin"),
				action.Execute(withCwd(options.Path), "git", "checkout", options.CommitHash),
				action.Execute(withCwd(options.Path), "git", "clean", "-xfd"),
				action.Execute(
					withCwd(options.Path),
					"git",
					"submodule",
					"foreach",
					"--recursive",
					"git",
					"clean",
					"-xfd",
				),
				installAction,
				action.Func(fmt.Sprintf("Update %s", shortPath), func() error {
					var stderr, stdout bytes.Buffer
					err := exec.Command().
						WithCwd(options.Path).
						WithBufout(&stdout, &stderr).
						Args("git", "rev-parse", "HEAD").Run()
					if err != nil {
						return err
					}
					if file.Exists(hashFile) {
						if err := os.Remove(hashFile); err != nil {
							return err
						}
					}
					currentHash := strings.Trim(stdout.String(), "\n ")
					return os.WriteFile(
						hashFile,
						[]byte(currentHash),
						0o644,
					)
				}),
			},
		},
	}
}

func cmd() *exec.Cmd {
	return exec.Command().WithStdio()
}

func getCurrentHashFromFile(hashFile string) string {
	if !file.Exists(hashFile) {
		return ""
	}
	file, readErr := os.ReadFile(hashFile)
	if readErr != nil {
		return ""
	}
	return strings.Trim(string(file), "\n ")
}

func getCurrentHashFromRepo(dir string) (string, error) {
	var stderr, stdout bytes.Buffer
	err := exec.Command().
		WithCwd(dir).
		WithBufout(&stdout, &stderr).
		Args("git", "rev-parse", "HEAD").Run()
	if err != nil {
		return "", err
	}
	return strings.Trim(stdout.String(), "\n "), nil
}

func updateHash(hashFile string, hash string) error {
	if file.Exists(hashFile) {
		if err := os.Remove(hashFile); err != nil {
			return err
		}
	}
	return os.WriteFile(
		hashFile,
		[]byte(hash),
		0o644,
	)
}

func InstallFromRepo(ctx context.Context, options RepoInstallOptions, installFn func(ctx context.Context) error) error {
	installPrefix := ctx.FromHome(".local")
	hashFileName := fmt.Sprintf(".%s.hash", options.Name)
	hashFile := path.Join(installPrefix, hashFileName)

	if !file.Exists(options.Path) {
		err := exec.RunAll(
			cmd().Args("mkdir", "-p", path.Dir(options.Path)),
			cmd().Args("git", "clone", options.RepoUrl, options.Path),
		)
		if err != nil {
			return err
		}
	}

	currentHash := getCurrentHashFromFile(hashFile)

	if currentHash != options.CommitHash {
		cmdErr := exec.RunAll(
			cmd().Args("mkdir", "-p", installPrefix),
			cmd().WithCwd(options.Path).Args("git", "fetch", "origin"),
			cmd().WithCwd(options.Path).Args("git", "checkout", options.CommitHash),
			cmd().WithCwd(options.Path).Args("git", "clean", "-xfd"),
			cmd().WithCwd(options.Path).Args(
				"git", "submodule", "foreach", "--recursive", "git", "clean", "-xfd",
			),
		)
		if cmdErr != nil {
			return cmdErr
		}
		if err := installFn(ctx); err != nil {
			return err
		}

		newHash, newHashErr := getCurrentHashFromRepo(options.Path)
		if newHashErr != nil {
			return newHashErr
		}
		if err := updateHash(hashFile, newHash); err != nil {
			return err
		}
	}
	return nil
}
