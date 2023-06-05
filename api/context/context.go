package context

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/api/platform"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/config"
	"github.com/wkozyra95/dotfiles/utils/file"
)

type Context struct {
	PkgInstaller      api.PackageInstaller
	Username          string
	Homedir           string
	Environment       string
	EnvironmentConfig env.EnvironmentConfig
}

func CreateContext() Context {
	userInfo, userErr := user.Current()
	if userErr != nil {
		panic(userErr)
	}
	homedir, homedirErr := os.UserHomeDir()
	if homedirErr != nil {
		panic(homedirErr)
	}
	pkgInstaller, pkgInstallerErr := platform.GetPackageManager()
	if pkgInstallerErr != nil {
		panic(pkgInstallerErr)
	}
	environment, environmentErr := ensureEnvironmentConfigured(homedir)
	if environmentErr != nil {
		panic(environmentErr)
	}

	return Context{
		Username:          userInfo.Username,
		Homedir:           homedir,
		Environment:       environment,
		EnvironmentConfig: config.GetConfig(),
		PkgInstaller:      pkgInstaller,
	}
}

func CreateContextForEnvironment(environment string) Context {
	userInfo, userErr := user.Current()
	if userErr != nil {
		panic(userErr)
	}
	homedir, homedirErr := os.UserHomeDir()
	if homedirErr != nil {
		panic(homedirErr)
	}
	pkgInstaller, pkgInstallerErr := platform.GetPackageManager()
	if pkgInstallerErr != nil {
		panic(pkgInstallerErr)
	}
	environmentErr := ensureSpecificEnvironment(homedir, environment)
	if environmentErr != nil {
		panic(environmentErr)
	}

	return Context{
		Username:          userInfo.Username,
		Homedir:           homedir,
		Environment:       environment,
		EnvironmentConfig: config.GetConfig(),
		PkgInstaller:      pkgInstaller,
	}
}

func (c Context) FromHome(relative string) string {
	return path.Join(c.Homedir, relative)
}

func (c Context) FromEnvDir(relative string) string {
	return path.Join(c.Homedir, ".dotfiles/env", c.Environment, relative)
}

func ensureSpecificEnvironment(homedir string, environment string) error {
	currentEnvironment := os.Getenv("CURRENT_ENV")
	if currentEnvironment == environment {
		return nil
	}
	ensureConfigErr := file.EnsureTextWithRegexp(
		path.Join(homedir, ".zshrc.local"),
		fmt.Sprintf("export CURRENT_ENV=\"%s\"", environment),
		regexp.MustCompile("export CURRENT_ENV.*"),
	)
	os.Setenv("CURRENT_ENV", environment)
	return ensureConfigErr
}

func ensureEnvironmentConfigured(homedir string) (string, error) {
	environment := os.Getenv("CURRENT_ENV")
	if environment != "" {
		return environment, nil
	}
	response, responseErr := (&promptui.Prompt{Label: "Specify name of this environment"}).Run()
	if responseErr != nil {
		return "", responseErr
	}
	ensureConfigErr := file.EnsureTextWithRegexp(
		path.Join(homedir, ".zshrc.local"),
		fmt.Sprintf("export CURRENT_ENV=\"%s\"", response),
		regexp.MustCompile("export CURRENT_ENV.*"),
	)
	os.Setenv("CURRENT_ENV", response)
	return response, ensureConfigErr
}
