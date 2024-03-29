package context

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/env/config"
	"github.com/wkozyra95/dotfiles/utils/file"
)

type Context struct {
	Username          string
	Group             string
	Homedir           string
	Environment       string
	EnvironmentConfig env.EnvironmentConfig
}

func CreateContext() Context {
	userInfo, userErr := user.Current()
	if userErr != nil {
		panic(userErr)
	}
	groupInfo, groupErr := user.LookupGroupId(userInfo.Gid)
	if groupErr != nil {
		panic(groupErr)
	}
	homedir, homedirErr := os.UserHomeDir()
	if homedirErr != nil {
		panic(homedirErr)
	}
	environment, environmentErr := ensureEnvironmentConfigured(homedir)
	if environmentErr != nil {
		panic(environmentErr)
	}

	return Context{
		Username:          userInfo.Username,
		Group:             groupInfo.Name,
		Homedir:           homedir,
		Environment:       environment,
		EnvironmentConfig: config.GetConfig(),
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
	environmentErr := ensureSpecificEnvironment(homedir, environment)
	if environmentErr != nil {
		panic(environmentErr)
	}

	return Context{
		Username:          userInfo.Username,
		Homedir:           homedir,
		Environment:       environment,
		EnvironmentConfig: config.GetConfig(),
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
