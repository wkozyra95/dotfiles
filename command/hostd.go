package command

import (
	"fmt"
	"os"
	"time"

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
		Short: "HTTP server for managing this host remotely (status / suspend / power off / files / notify / push tokens)",
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

	var stateStateDir string
	stateCmd := &cobra.Command{
		Use:   "state",
		Short: "Print what the service stores: the token used to sign requests and the registered push devices (needs root)",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := printState(stateStateDir); err != nil {
				log.Error(err)
				os.Exit(1)
			}
		},
	}
	stateCmd.Flags().StringVar(&stateStateDir, "state-dir", stateDir, "state directory of the service")

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
	hostdCmd.AddCommand(stateCmd)
	hostdCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(hostdCmd)
}

// printState shows the token and the push devices, whichever of the two can be
// read, and fails if either can not.
func printState(stateDir string) error {
	token, tokenErr := hostd.ReadToken(stateDir)
	if tokenErr != nil {
		fmt.Println("Token: not available, it is generated when the hostd service starts")
	} else {
		fmt.Printf("Token: %s\n", token)
	}

	devices, devicesErr := hostd.NewPushTokenStore(stateDir).List()
	switch {
	case devicesErr != nil:
		fmt.Println("Push devices: not available")
	case len(devices) == 0:
		fmt.Println("Push devices: none, the app registers with POST /push-token")
	default:
		fmt.Printf("Push devices (%d):\n", len(devices))
		for _, device := range devices {
			fmt.Printf(
				"  %-24s registered %s  token %s\n",
				device.Name,
				time.Unix(device.RegisteredAt, 0).Format("2006-01-02 15:04"),
				abbreviate(device.Token),
			)
		}
	}

	if tokenErr != nil {
		return fmt.Errorf("unable to read the token [%w]", tokenErr)
	}
	if devicesErr != nil {
		return fmt.Errorf("unable to read the push devices [%w]", devicesErr)
	}
	return nil
}

// abbreviate keeps enough of an FCM token to tell devices apart, the whole
// thing is 150+ characters.
func abbreviate(token string) string {
	const keep = 12
	if len(token) <= 2*keep+1 {
		return token
	}
	return token[:keep] + "…" + token[len(token)-keep:]
}
