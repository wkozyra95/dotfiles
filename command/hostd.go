package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/api/hostd"
	"github.com/wkozyra95/dotfiles/api/hostd/server"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func RegisterHostdCmds(rootCmd *cobra.Command) {
	hostdCmd := &cobra.Command{
		Use:   "hostd",
		Short: "HTTP server for managing this host remotely (status / suspend / power off / files / notify)",
	}

	// systemd passes it for StateDirectory=
	stateDir := envOrDefault("STATE_DIRECTORY", hostd.DefaultStateDir)

	config := server.Config{}
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the server, run by the hostd systemd service (nix/nix-modules/hostd.nix)",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := server.Serve(config); err != nil {
				log.Error(err)
				os.Exit(1)
			}
		},
	}
	serveCmd.Flags().StringVar(&config.Listen, "listen", server.DefaultListen, "address to listen on")
	serveCmd.Flags().StringVar(&config.StateDir, "state-dir", stateDir, "directory with the token")
	serveCmd.Flags().
		StringVar(&config.FilesDir, "files-dir", "", "enable /files endpoints, directory for file transfers")
	serveCmd.Flags().
		BoolVar(&config.Notify, "notify", false, "enable /notify endpoint, desktop notifications in the session of the user running the service")

	var tokenStateDir string
	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Print the token used to sign requests (needs root)",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			token, err := hostd.ReadToken(tokenStateDir)
			if err != nil {
				log.Errorf("Unable to read the token, it is generated when hostd service starts [%v]", err)
				os.Exit(1)
			}
			fmt.Println(token)
		},
	}
	tokenCmd.Flags().StringVar(&tokenStateDir, "state-dir", stateDir, "directory with the token")

	var logsLines int
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "Stream logs of the running hostd systemd service (journalctl -f)",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			journalctl, err := exec.Command().WithStdio().Args(
				"journalctl", "--unit", "hostd", "--follow", "--lines", fmt.Sprint(logsLines),
			).Start()
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			// Ctrl-C reaches journalctl as well, that is the normal way to stop
			if err := journalctl.Wait(); err != nil && journalctl.ProcessState.ExitCode() != -1 {
				os.Exit(journalctl.ProcessState.ExitCode())
			}
		},
	}
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 100, "number of past log lines to show before following")

	hostdCmd.AddCommand(serveCmd)
	hostdCmd.AddCommand(tokenCmd)
	hostdCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(hostdCmd)
}
