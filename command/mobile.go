package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/api/hostd"
	"github.com/wkozyra95/dotfiles/api/mobile"
	"github.com/wkozyra95/dotfiles/utils/menu"
	"github.com/wkozyra95/dotfiles/utils/notify"
)

const (
	allDevices = "All devices"
	// Gitignored, the key of a service account that only has the "Firebase
	// Cloud Messaging API Admin" role.
	defaultServiceAccountFile = ".dotfiles/secrets/fcm-service-account.json"
)

func RegisterMobileCmds(rootCmd *cobra.Command) {
	mobileCmd := &cobra.Command{
		Use:   "mobile",
		Short: "Push notifications to the phones registered with hostd on this host",
	}

	var stateDir, device, serviceAccountFile string
	sendCmd := &cobra.Command{
		Use:   "send [text]",
		Short: "Send text as a push notification, without arguments fuzzel asks for the device and the text",
		Long: "Send text as a push notification to the devices registered with `POST /push-token` of the hostd\n" +
			"service running on this host (reads its state directory, so it works where hostd runs as your user).\n" +
			"Without arguments fuzzel asks which device and what to send, meant for a sway keybinding.\n" +
			"With the text given it goes to all devices unless --device is set.\n" +
			"Delivery goes through FCM with the service account key from --service-account (the JSON\n" +
			"downloaded from the Firebase console, the file is gitignored).",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			interactive := len(args) == 0
			text := ""
			if !interactive {
				text = args[0]
			}
			if err := mobileSend(stateDir, serviceAccountFile, device, text, interactive); err != nil {
				if interactive {
					notify.Notify(
						notify.Notification{Title: "Push failed", Message: err.Error(), Urgency: notify.Critical},
					)
				}
				log.Error(err)
				os.Exit(1)
			}
		},
	}
	sendCmd.Flags().
		StringVar(&stateDir, "state-dir", envOrDefault("STATE_DIRECTORY", hostd.DefaultStateDir), "state directory of hostd with the registered tokens")
	sendCmd.Flags().StringVar(&device, "device", "", "send only to the device registered under this name")
	sendCmd.Flags().StringVar(
		&serviceAccountFile, "service-account", defaultServiceAccountPath(), "Firebase service account key (JSON)",
	)

	mobileCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(mobileCmd)
}

func mobileSend(stateDir string, serviceAccountFile string, deviceName string, text string, interactive bool) error {
	store := hostd.NewPushTokenStore(stateDir)
	devices, err := store.List()
	if errors.Is(err, os.ErrPermission) {
		return fmt.Errorf(
			"%w (hostd has to run as your user for `mycli mobile send`, see nix/nix-modules/hostd.nix)",
			err,
		)
	}
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return errors.New("no devices registered, the app has to register its token with POST /push-token first")
	}

	if interactive {
		// a cancelled prompt is not an error, nothing is sent
		if deviceName, text = promptDeviceAndText(devices, deviceName); text == "" {
			return nil
		}
	}
	targets := devices
	if deviceName != "" {
		targets = nil
		for _, candidate := range devices {
			if candidate.Name == deviceName {
				targets = append(targets, candidate)
			}
		}
		if len(targets) == 0 {
			return fmt.Errorf(
				"device %q is not registered, registered: %s",
				deviceName,
				strings.Join(deviceNames(devices), ", "),
			)
		}
	}

	account, err := fcmServiceAccount(serviceAccountFile)
	if err != nil {
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	ctx := context.Background()
	client := mobile.NewClient(ctx, account)
	sent := []string{}
	failed := []string{}
	for _, target := range targets {
		err := client.SendText(ctx, target.Token, hostname, text)
		if errors.Is(err, mobile.ErrUnregistered) {
			log.Warnf("%s: %v, removing its token", target.Name, err)
			if err := store.Unregister(target.Token); err != nil {
				log.Warnf("Failed to remove the token of %s [%v]", target.Name, err)
			}
		} else if err != nil {
			log.Errorf("%s: %v", target.Name, err)
		}
		if err != nil {
			failed = append(failed, target.Name)
		} else {
			sent = append(sent, target.Name)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("sending to %s failed", strings.Join(failed, ", "))
	}
	summary := fmt.Sprintf("Sent to %s", strings.Join(sent, ", "))
	log.Info(summary)
	if interactive {
		notify.Notify(notify.Notification{Title: summary, Message: text, Urgency: notify.Low})
	}
	return nil
}

// promptDeviceAndText asks with fuzzel, an empty text means cancelled. The
// device selection is skipped with a single device or a device given with
// --device, "" means all of them.
func promptDeviceAndText(devices []hostd.PushDevice, deviceName string) (string, string) {
	if deviceName == "" && len(devices) > 1 {
		selected, ok := menu.Select("Send to", append(deviceNames(devices), allDevices))
		if !ok {
			return "", ""
		}
		if selected != allDevices {
			deviceName = selected
		}
	}
	prompt := "Send to " + allDevices
	if deviceName != "" {
		prompt = "Send to " + deviceName
	} else if len(devices) == 1 {
		prompt = "Send to " + devices[0].Name
	}
	return deviceName, menu.Prompt(prompt)
}

func deviceNames(devices []hostd.PushDevice) []string {
	names := []string{}
	for _, device := range devices {
		names = append(names, device.Name)
	}
	return names
}

func defaultServiceAccountPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return defaultServiceAccountFile
	}
	return path.Join(homedir, defaultServiceAccountFile)
}

func fcmServiceAccount(file string) (mobile.ServiceAccount, error) {
	content, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return mobile.ServiceAccount{}, fmt.Errorf(
			"%s does not exist, download the key of the service account from the Firebase console", file,
		)
	}
	if err != nil {
		return mobile.ServiceAccount{}, err
	}
	account, err := mobile.ParseServiceAccount(content)
	if err != nil {
		return mobile.ServiceAccount{}, fmt.Errorf("%s: %w", file, err)
	}
	return account, nil
}
