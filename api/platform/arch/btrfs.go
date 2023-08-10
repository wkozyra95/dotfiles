package arch

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strings"

	a "github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/logger"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

var log = logger.NamedLogger("btrfs")

type Volume struct {
	Name       string
	Path       string
	ID         string
	ParrentID  string
	TopLevelID string
}

func parseVolumeList(stdout string) ([]Volume, error) {
	lines := strings.Split(strings.Trim(stdout, " \n"), "\n")
	rg := regexp.MustCompile("ID ([0-9]+) gen ([0-9]+) parent ([0-9]+) top level ([0-9]+) path (.+)")
	snapshots := []Volume{}
	for _, line := range lines {
		match := rg.FindStringSubmatch(line)
		if len(match) == 0 {
			return nil, fmt.Errorf("No match for line '%s'", line)
		}
		snapshots = append(snapshots, Volume{
			Name:       match[5],
			Path:       match[5],
			ID:         match[1],
			ParrentID:  match[3],
			TopLevelID: match[4],
		})
	}
	return snapshots, nil
}

func getSubvolumeId(path string) (string, error) {
	var stdout bytes.Buffer
	cmdErr := exec.Command().
		WithBufout(&stdout, &bytes.Buffer{}).
		Run("sudo", "btrfs", "subvolume", "show", path)
	if cmdErr != nil {
		return "", cmdErr
	}
	rg := regexp.MustCompile(`Subvolume ID:\s+([0-9]+)`)
	match := rg.FindStringSubmatch(stdout.String())
	if len(match) < 2 {
		return "", fmt.Errorf("no match for subvolume id\n%s", stdout.String())
	}
	return match[1], nil
}

func getSubvolumeChildren(subvolID string) ([]Volume, error) {
	var stdout bytes.Buffer
	cmdErr := exec.Command().
		WithBufout(&stdout, &bytes.Buffer{}).
		Run("sudo", "btrfs", "subvolume", "list", "-p", "/")
	if cmdErr != nil {
		return nil, cmdErr
	}
	subvol, subvolErr := parseVolumeList(stdout.String())
	if subvolErr != nil {
		return nil, subvolErr
	}
	filterSubvol := []Volume{}
	for _, subvol := range subvol {
		if subvolID == subvol.ParrentID {
			filterSubvol = append(filterSubvol, subvol)
		}
	}
	return filterSubvol, nil
}

func GetSnapshots() ([]Volume, error) {
	var snaphotListRaw bytes.Buffer
	cmdErr := exec.Command().
		WithBufout(&snaphotListRaw, &bytes.Buffer{}).
		Run("sudo", "btrfs", "subvolume", "list", "-p", "-o", "/run/btrfs-root/__snapshot/")
	if cmdErr != nil {
		return nil, cmdErr
	}
	snapshots, snapshotsErr := parseVolumeList(snaphotListRaw.String())
	if snapshotsErr != nil {
		return nil, snapshotsErr
	}
	filterSnapshots := []Volume{}
	for _, snapshot := range snapshots {
		if strings.HasPrefix(snapshot.Path, "__snapshot") {
			filterSnapshots = append(filterSnapshots, snapshot)
		}
	}
	return filterSnapshots, nil
}

func SelectSnapshot() (Volume, error) {
	snapshots, snapshotErr := GetSnapshots()
	if snapshotErr != nil {
		return Volume{}, snapshotErr
	}
	snapshot, isSelected := prompt.SelectPrompt(
		"Select snapshot",
		snapshots,
		func(s Volume) string { return s.Name },
	)
	if !isSelected {
		return snapshot, fmt.Errorf("No snapshots were selected")
	}
	return snapshot, nil
}

func RestoreRootSnapshot() error {
	snapshot, snapshotErr := SelectSnapshot()
	if snapshotErr != nil {
		return snapshotErr
	}
	rootPartition := "/run/btrfs-root/__current/root"
	rootPartitionBackup := "/run/btrfs-root/__current/root-tmp"
	actions := a.List{
		a.WithCondition{
			If:   a.PathExists(rootPartitionBackup),
			Then: a.ShellCommand("sudo", "btrfs", "subvolume", "delete", rootPartitionBackup),
		},
		a.WithCondition{
			If:   a.PathExists(rootPartition),
			Then: a.ShellCommand("sudo", "mv", rootPartition, rootPartitionBackup),
		},
		a.ShellCommand(
			"sudo",
			"btrfs",
			"subvolume",
			"snapshot",
			path.Join("/run/btrfs-root/", snapshot.Path),
			rootPartition,
		),
		a.Func("copy subvolume children", func() error {
			tmpRootId, tmpRootIdErr := getSubvolumeId(rootPartitionBackup)
			if tmpRootIdErr != nil {
				return tmpRootIdErr
			}
			childVolumes, childVolumesErr := getSubvolumeChildren(tmpRootId)
			if childVolumesErr != nil {
				return childVolumesErr
			}
			for _, child := range childVolumes {
				pathSuffix := strings.TrimPrefix(child.Path, "__current/root-tmp/")
				destinationPath := path.Join(rootPartition, pathSuffix)
				srcPath := path.Join("/run/btrfs-root", child.Path)
				fmt.Printf("reattach volume %s -> %s\n", srcPath, destinationPath)
				if err := exec.Command().WithStdio().Run("sudo", "rmdir", destinationPath); err != nil {
					return err
				}
				err := exec.Command().WithStdio().Run(
					"sudo", "mv",
					srcPath,
					destinationPath,
				)
				if err != nil {
					return err
				}
			}
			return nil
		}),
		a.ShellCommand("sudo", "btrfs", "subvolume", "delete", rootPartitionBackup),
	}
	return a.RunActions(actions)
}

func CleanupSnapshots() error {
	snapshots, snapshotErr := GetSnapshots()
	if snapshotErr != nil {
		return snapshotErr
	}
	selectedSnapshots := prompt.MultiselectPrompt(
		"Delete snapshot",
		snapshots,
		func(s Volume) string { return s.Name },
	)
	log.Info("")
	log.Info("Deleting snapshots")
	for _, snapshot := range selectedSnapshots {
		fullPath := path.Join("/run/btrfs-root", snapshot.Path)
		log.Infof(" - %s", fullPath)
		err := exec.Command().
			WithSudo().
			Run("btrfs", "subvolume", "delete", fullPath)
		if err != nil {
			return err
		}
	}
	return nil
}
