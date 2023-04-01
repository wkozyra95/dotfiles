package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var example = `  feature-1 4ea9770 [origin/feature-1] test commit
  feature-2 4ea9770 test commit
  feature-6 4ea9770 [origin/feature-6] test commit
  feature-7 4ea9770 [origin/feature-7: gone] test commit
* main      4ea9770 [origin/main] test commit
`

func TestGetGitBranchInfo(t *testing.T) {
	result := parseListBranches(example)
	assert.Equal(t, []BranchInfo{
		{
			IsCurrent:    false,
			Name:         "feature-1",
			CommitHash:   "4ea9770",
			RemoteBranch: "origin/feature-1",
			IsRemoteGone: false,
			Message:      "test commit",
		},
		{
			IsCurrent:    false,
			Name:         "feature-2",
			CommitHash:   "4ea9770",
			RemoteBranch: "",
			IsRemoteGone: false,
			Message:      "test commit",
		},
		{
			IsCurrent:    false,
			Name:         "feature-6",
			CommitHash:   "4ea9770",
			RemoteBranch: "origin/feature-6",
			IsRemoteGone: false,
			Message:      "test commit",
		},
		{
			IsCurrent:    false,
			Name:         "feature-7",
			CommitHash:   "4ea9770",
			RemoteBranch: "origin/feature-7",
			IsRemoteGone: true,
			Message:      "test commit",
		},
		{
			IsCurrent:    true,
			Name:         "main",
			CommitHash:   "4ea9770",
			RemoteBranch: "origin/main",
			IsRemoteGone: false,
			Message:      "test commit",
		},
	}, result)
}
