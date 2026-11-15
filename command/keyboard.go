package command

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/keyboard"
	"github.com/wkozyra95/dotfiles/utils/notify"
)

func RegisterKeyboardCmds(rootCmd *cobra.Command) {
	keyboardCmd := &cobra.Command{
		Use:   "keyboard",
		Short: "keyboard used with this environment",
	}

	previewCmd := &cobra.Command{
		Use:   "preview",
		Short: "Show the keymap of the keyboard, a second call closes the preview",
		Long: "Draw the layers of the keymap of the keyboard used with this environment and show them in a\n" +
			"floating window, q or Escape closes it. Calling it again closes the preview too, so a single\n" +
			"sway keybinding toggles it.\n" +
			"The keyboards and their environments are defined in api/keyboard, the firmware is not read.",
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.CreateContext()
			if err := keyboard.Preview(ctx); err != nil {
				// started from the launcher or a keybinding nobody sees the log
				notify.Notify(
					notify.Notification{
						Title:   "Keyboard preview failed",
						Message: err.Error(),
						Urgency: notify.Critical,
					},
				)
				log.Error(err)
				os.Exit(1)
			}
		},
	}

	keyboardCmd.AddCommand(previewCmd)
	rootCmd.AddCommand(keyboardCmd)
}
