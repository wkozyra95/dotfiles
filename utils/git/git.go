package git

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

var branchInfoRg = regexp.MustCompile(
	`(?m)^(\*?)\s+(\S+)\s+(\S+)\s+(\[(\S+)(: gone|)\]\s+|)(.+)$`,
)

type BranchInfo struct {
	IsCurrent    bool
	Name         string
	CommitHash   string
	RemoteBranch string
	IsRemoteGone bool
	Message      string
}

func (b BranchInfo) String() string {
	maybeRemote := ""
	if b.RemoteBranch != "" {
		maybeRemote = fmt.Sprintf(" [%s]", b.RemoteBranch)
	}
	maybeGone := ""
	if b.IsRemoteGone {
		maybeGone = "[gone] "
	}
	return fmt.Sprintf("%s%s%s - %s %s", maybeGone, b.Name, maybeRemote, b.CommitHash, b.Message)
}

func Prune() error {
	return exec.Command().WithStdio().Run("git", "fetch", "origin", "--prune")
}

func DeleteBranch(name string) error {
	return exec.Command().WithStdio().Run("git", "branch", "-D", name)
}

func ListBranches() ([]BranchInfo, error) {
	var stdout bytes.Buffer
	spawnErr := exec.Command().WithBufout(&stdout, &bytes.Buffer{}).Run("git", "branch", "-vv")
	if spawnErr != nil {
		return nil, spawnErr
	}
	return parseListBranches(stdout.String()), nil
}

func parseListBranches(gitOutput string) []BranchInfo {
	matches := branchInfoRg.FindAllStringSubmatch(gitOutput, -1)
	results := make([]BranchInfo, 0, len(matches))
	for _, match := range matches {
		branch := parseBranchInfo(match)
		results = append(results, branch)
	}
	return results
}

func parseBranchInfo(match []string) BranchInfo {
	return BranchInfo{
		IsCurrent:    match[1] == "*",
		Name:         match[2],
		CommitHash:   match[3],
		RemoteBranch: match[5],
		IsRemoteGone: match[6] == ": gone",
		Message:      match[7],
	}
}
