package command

import (
	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/utils/fn"
	"github.com/wkozyra95/dotfiles/utils/git"
	"github.com/wkozyra95/dotfiles/utils/prompt"
)

func RegisterGitCmds(rootCmd *cobra.Command) {
	gitCmds := &cobra.Command{
		Use:   "git",
		Short: "git helper",
	}

	gitPrune := &cobra.Command{
		Use:   "prune",
		Short: "better git prune",
		Run: func(cmd *cobra.Command, args []string) {
			if err := git.Prune(); err != nil {
				log.Error(err.Error())
			}
			branches, branchesErr := git.ListBranches()
			if branchesErr != nil {
				log.Error(branchesErr)
			}
			branchesForDeletion := fn.Filter(branches, func(b git.BranchInfo) bool {
				return (b.IsRemoteGone || b.RemoteBranch == "") && !b.IsCurrent && b.Name != "main" &&
					b.Name != "master"
			})
			selctedBranches := prompt.MultiselectPrompt(
				"Select which branches you want to delete",
				branchesForDeletion,
				git.BranchInfo.String,
			)
			for _, branch := range selctedBranches {
				if err := git.DeleteBranch(branch.Name); err != nil {
					log.Error(err)
				}
			}
		},
	}

	gitCmds.AddCommand(gitPrune)

	rootCmd.AddCommand(gitCmds)
}
