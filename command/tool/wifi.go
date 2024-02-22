package tool

import (
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/wkozyra95/dotfiles/secret"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

func wifiManager() {
	homedir, homedirErr := os.UserHomeDir()
	if homedirErr != nil {
		panic(homedirErr)
	}
	secrets := secret.BestEffortReadSecrets(homedir)
	if secrets == nil {
		return
	}
	networks := []struct {
		ssid     string
		password string
	}{
		{secrets.Wifi.HomeUpc.Ssid, secrets.Wifi.HomeUpc.Password},
		{secrets.Wifi.HomeOrange.Ssid, secrets.Wifi.HomeOrange.Password},
	}
	networkNames := make([]string, len(networks))
	for i, network := range networks {
		networkNames[i] = network.ssid
	}
	index, _, selectErr := (&promptui.Select{
		Label: "Select preconfigured network",
		Items: networkNames,
	}).Run()
	if selectErr != nil {
		panic(selectErr)
	}
	if exec.CommandExists("nmcli") {
		err := exec.Command().
			WithStdio().
			Args("nmcli", "dev", "wifi", "con", networks[index].ssid, "password", networks[index].password).Run()
		if err != nil {
			log.Error(err)
		}
	} else {
		log.Infof("Unsupported auto-connect (ssid: %s, password: %s)", networks[index].ssid, networks[index].password)
	}
}

func registerWifiCommands() *cobra.Command {
	driveCmd := &cobra.Command{
		Use:   "wifi",
		Short: "wifi managment",
		Run: func(cmd *cobra.Command, args []string) {
			wifiManager()
		},
	}

	return driveCmd
}
