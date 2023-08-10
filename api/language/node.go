package language

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

var tsconfig = map[string]any{
	"compilerOptions": map[string]any{
		"target":                     "es2020",
		"lib":                        []string{"es2020"},
		"module":                     "commonjs",
		"sourceMap":                  true,
		"inlineSources":              true,
		"strict":                     true,
		"noUnusedLocals":             true,
		"noUnusedParameters":         true,
		"noImplicitReturns":          true,
		"noFallthroughCasesInSwitch": true,
		"esModuleInterop":            true,
		"skipLibCheck":               true,
		"noEmit":                     true,
	},
	"include": []string{
		"index.ts",
	},
}

func NodePackageInstallAction(pkg string, reinstallCond action.Condition) action.Object {
	return action.WithCondition{
		If: action.CommandExists("volta"),
		Then: action.WithCondition{
			If: action.Or(
				action.FuncCond(fmt.Sprintf("is not %s installed", pkg), func() (bool, error) {
					ok, err := isVoltaPackageInstalled(pkg)
					return !ok, err
				}),
				reinstallCond,
			),
			Then: action.ShellCommand("volta", "install", pkg),
		},
		Else: action.WithCondition{
			If: action.Or(
				action.FuncCond(fmt.Sprintf("is not %s installed", pkg), func() (bool, error) {
					ok, err := isGlobalNpmPackageInstalled(pkg)
					return !ok, err
				}),
				reinstallCond,
			),
			Then: action.ShellCommand("npm", "-g", "install", pkg),
		},
	}
}

func isVoltaPackageInstalled(pkg string) (bool, error) {
	var stdout bytes.Buffer
	if err := exec.Command().WithBufout(&stdout, &bytes.Buffer{}).Run("volta", "list", "--format", "plain"); err != nil {
		return false, err
	}
	isInstalled, err := regexp.MatchString(regexp.QuoteMeta(fmt.Sprintf("package %s", pkg)), stdout.String())
	if err != nil {
		return false, err
	}
	return isInstalled, nil
}

func isGlobalNpmPackageInstalled(pkg string) (bool, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exec.Command().WithBufout(&stdout, &stderr).Run("npm", "list", "-g", "--json")
	var parsedJson struct {
		Dependencies map[string]interface{} `json:"dependencies"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &parsedJson); err != nil {
		return false, err
	}
	for key := range parsedJson.Dependencies {
		if key == pkg {
			return true, nil
		}
	}
	return false, nil
}

func NodePlaygroundCreate(playgroundPath string) error {
	if file.Exists(playgroundPath) {
		return nil
	}
	if err := exec.Command().Run("mkdir", "-p", playgroundPath); err != nil {
		return err
	}
	if err := exec.Command().Run("touch", path.Join(playgroundPath, "index.ts")); err != nil {
		return err
	}
	if err := exec.Command().WithCwd(playgroundPath).Run("npm", "init", "-y"); err != nil {
		return err
	}
	rawTsconfigJson, marshalErr := json.Marshal(tsconfig)
	if marshalErr != nil {
		return marshalErr
	}
	if err := os.WriteFile(path.Join(playgroundPath, "tsconfig.json"), rawTsconfigJson, 0o644); err != nil {
		return err
	}
	yarnErr := exec.Command().
		WithCwd(playgroundPath).
		Run("yarn", "add", "--dev", "@types/node@14", "eslint", "eslint-plugin-import", "prettier", "typescript", "ts-node")
	if yarnErr != nil {
		return yarnErr
	}
	return nil
}

func NodePlaygroundDelete(playgroundPath string) error {
	if !file.Exists(playgroundPath) {
		return nil
	}
	if err := os.RemoveAll(playgroundPath); err != nil {
		return err
	}
	return nil
}

func NodePlaygroundNodeShell(playgroundPath string) error {
	return api.AlacrittyCall(
		api.AlacrittyConfig{Command: "./node_modules/.bin/ts-node", Cwd: playgroundPath, ShouldRetry: true},
	)
}

func NodePlaygroundZshShell(playgroundPath string) error {
	return api.AlacrittyCall(
		api.AlacrittyConfig{Command: "zsh", Cwd: playgroundPath, ShouldRetry: false},
	)
}

func NodePlaygroundInstall(playgroundPath string, pkg string) error {
	if err := exec.Command().WithCwd(playgroundPath).Run("yarn", "add", pkg); err != nil {
		return err
	}
	splitPackage := strings.Split(pkg, "/")
	if splitPackage[0] == "@types" {
		return nil
	}
	// without @npmorg prefix
	sanitizedName := splitPackage[len(splitPackage)-1]
	splitNameSanitizedName := strings.Split(sanitizedName, "-")
	if len(splitNameSanitizedName) > 1 {
		for i, element := range splitNameSanitizedName {
			if i >= 1 {
				splitNameSanitizedName[i] = strings.Title(element)
			}
		}
	}
	// convert - to camel case
	sanitizedName = strings.Join(splitNameSanitizedName, "")

	ensureErr := file.EnsureTextWithRegexp(
		path.Join(playgroundPath, "index.ts"),
		fmt.Sprintf("import %s from \"%s\";", sanitizedName, pkg),
		regexp.MustCompile(fmt.Sprintf(".*%s.*", regexp.QuoteMeta(fmt.Sprintf("from \"%s\"", pkg)))),
	)
	if ensureErr != nil {
		return ensureErr
	}
	return nil
}
