package command

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wkozyra95/dotfiles/action"
	"github.com/wkozyra95/dotfiles/api/context"
)

func RegisterCompletionCmds(rootCmd *cobra.Command) {
	completionCmd := &cobra.Command{
		Use:                   "completion",
		Short:                 "Generate completion script",
		Long:                  "To load completions",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			err := action.Run(action.List{
				action.ShellCommand("mkdir", "-p", ctx.FromHome(".cache/mycli/completion/zsh")),
				action.EnsureText(
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
				),
				action.Func("Generate completion file", func() error {
					return cmd.Root().GenZshCompletionFile(ctx.FromHome(".cache/mycli/completion/zsh/_mycli"))
				}),
			})
			if err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(completionCmd)
}
