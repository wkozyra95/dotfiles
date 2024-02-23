package command

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

func RegisterCompletionCmds(rootCmd *cobra.Command) {
	generateCompletions := func(cmd *cobra.Command, ctx context.Context) error {
		if err := exec.Command().Args("mkdir", "-p", ctx.FromHome(".cache/mycli/completion/zsh")).Run(); err != nil {
			return err
		}
		textReplaceErr := file.EnsureTextWithRegexp(
			ctx.FromHome(".cache/mycli/completion/zsh_setup"),
			strings.Join(
				[]string{
					fmt.Sprintf("fpath=(%s \"${fpath[@]}\")", ctx.FromHome(".cache/mycli/completion/zsh")),
					"autoload -Uz compinit;",
					"compinit;",
					"",
				},
				"\n",
			),
			regexp.MustCompile("(?s).*"),
		)
		if textReplaceErr != nil {
			return textReplaceErr
		}
		return cmd.Root().GenZshCompletionFile(ctx.FromHome(".cache/mycli/completion/zsh/_mycli"))
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:                   "completion",
		Short:                 "Generate completion script",
		Long:                  "To load completions",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := generateCompletions(cmd, ctx); err != nil {
				panic(err)
			}
		},
	})
}
