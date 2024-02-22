package tool

import (
	gocontext "context"
	"os"
	"path"

	"github.com/google/go-github/v50/github"
	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"golang.org/x/oauth2"
)

const (
	releaseName = "release - v0.0.0"
	repoOwner   = "wkozyra95"
	repoName    = "dotfiles"
)

func registerReleaseCommands(rootCmd *cobra.Command) {
	upgradeCmd := &cobra.Command{
		Use:   "release",
		Short: "publish cli as a GitHub release",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			err := release(ctx)
			if err != nil {
				log.Error(err.Error())
			}
		},
	}

	rootCmd.AddCommand(upgradeCmd)
}

func release(ctx context.Context) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_RELEASE_TOKEN")},
	)
	tc := oauth2.NewClient(gocontext.Background(), ts)
	client := github.NewClient(tc)
	release, ensureReleaseErr := ensureRelease(client)
	if ensureReleaseErr != nil {
		return ensureReleaseErr
	}
	linuxBuildErr := exec.Command().WithStdio().
		WithCwd(path.Join(ctx.Homedir, ".dotfiles")).
		WithEnv("CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64").
		Args("go", "build", "-o", "bin/mycli-linux", ".").Run()

	if linuxBuildErr != nil {
		return linuxBuildErr
	}

	darwinBuildErr := exec.Command().WithStdio().
		WithCwd(path.Join(ctx.Homedir, ".dotfiles")).
		WithEnv("CGO_ENABLED=0", "GOOS=darwin", "GOARCH=amd64").
		Args("go", "build", "-o", "bin/mycli-darwin", ".").Run()
	if darwinBuildErr != nil {
		return darwinBuildErr
	}
	mycliLinux, openErr := os.Open(ctx.FromHome(".dotfiles/bin/mycli-linux"))
	if openErr != nil {
		return openErr
	}
	defer mycliLinux.Close()
	mycliDarwin, openErr := os.Open(ctx.FromHome(".dotfiles/bin/mycli-darwin"))
	if openErr != nil {
		return openErr
	}
	defer mycliDarwin.Close()

	_, _, updateErr := updateAssets(client, release.GetID(), mycliLinux, mycliDarwin)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

func updateAssets(
	client *github.Client,
	releaseID int64,
	mycliLinux, mycliDarwin *os.File,
) (*github.ReleaseAsset, *github.ReleaseAsset, error) {
	oldAssetList, _, getErr := client.Repositories.ListReleaseAssets(
		gocontext.Background(),
		repoOwner,
		repoName,
		releaseID,
		&github.ListOptions{},
	)
	if getErr != nil {
		return nil, nil, getErr
	}
	for _, asset := range oldAssetList {
		log.Debugf("Deleting asset %s", *asset.Name)
		if _, err := client.Repositories.DeleteReleaseAsset(gocontext.Background(), repoOwner, repoName, asset.GetID()); err != nil {
			return nil, nil, err
		}
	}
	linuxAsset, _, uploadLinuxErr := client.Repositories.UploadReleaseAsset(
		gocontext.Background(),
		repoOwner,
		repoName,
		releaseID,
		&github.UploadOptions{Name: "mycli-linux"},
		mycliLinux,
	)
	if uploadLinuxErr != nil {
		return nil, nil, uploadLinuxErr
	}

	darwinAsset, _, uploadDarwinErr := client.Repositories.UploadReleaseAsset(
		gocontext.Background(),
		repoOwner,
		repoName,
		releaseID,
		&github.UploadOptions{Name: "mycli-darwin"},
		mycliDarwin,
	)
	if uploadDarwinErr != nil {
		return nil, nil, uploadDarwinErr
	}
	return linuxAsset, darwinAsset, nil
}

func ensureRelease(client *github.Client) (*github.RepositoryRelease, error) {
	releaseName := releaseName
	tagName := "v0.0.0"
	existingRelease, findReleaseErr := findReleaseByName(client, releaseName)
	if findReleaseErr != nil {
		return nil, findReleaseErr
	}
	if existingRelease != nil {
		return existingRelease, nil
	}
	release, _, createErr := client.Repositories.CreateRelease(
		gocontext.Background(),
		repoOwner, repoName,
		&github.RepositoryRelease{Name: &releaseName, TagName: &tagName},
	)
	if createErr != nil {
		return nil, createErr
	}
	return release, nil
}

func findReleaseByName(client *github.Client, name string) (*github.RepositoryRelease, error) {
	releases, _, listErr := client.Repositories.ListReleases(
		gocontext.Background(),
		repoOwner,
		repoName,
		&github.ListOptions{},
	)
	if listErr != nil {
		return nil, listErr
	}
	for _, release := range releases {
		if *release.Name == name {
			return release, nil
		}
	}
	return nil, nil
}
