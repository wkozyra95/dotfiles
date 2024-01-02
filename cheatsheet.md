### Create hotspot
```bash
nmcli con add type wifi ifname wlan0 con-name Hostspot autoconnect yes ssid Hostspot
nmcli con modify Hostspot 802-11-wireless.mode ap 802-11-wireless.band bg ipv4.method shared
nmcli con modify Hostspot wifi-sec.key-mgmt wpa-psk
nmcli con modify Hostspot wifi-sec.psk "veryveryhardpassword1234"
nmcli con up Hostspot
```

### GPG

#### Setup signing commits

```bash
gpg --full-generate-key
gpg --list-secret-keys --keyid-format=long
gpg --armor --export 35DF8DFAD0E71E39F047BD01AE51A5682B78648C
```

#### Export and restore keyring

```bash
gpg --armor --export > public_keys.asc
gpg --armor --export-secret-keys > private_keys.asc
gpg --export-ownertrust > trustdb.txt
```

```bash
gpg --import public_keys.asc
gpg --import private_keys.asc
gpg --import-ownertrust trustdb.txt
```

### Partitioning

#### Encrypted usb drive
```bash
sudo sgdisk -Z /dev/sdb
sudo sgdisk -n 1::-0  --typecode=1:CA7D7CCB-63ED-4C53-861C-1742536059CC --change-name=1:'DRIVE_NAME' /dev/sdb
sudo cryptsetup luksFormat /dev/sdb1
sudo cryptsetup luksOpen /dev/sdb1 drive_name
sudo mkfs.ext4 /dev/mapper/drive_name -L drive_name
sudo mount /dev/mapper/drive_name /mnt/sdb1
```


### Nix settup

```
$ mkdir -p ~/.config/systemd/user/default.target.wants
$ ln -s /run/current-system/sw/lib/systemd/user/pipewire.service ~/.config/systemd/user/default.target.wants/
$ ln -s /run/current-system/sw/lib/systemd/user/wireplumber.service ~/.config/systemd/user/default.target.wants/
$ systemctl --user daemon-reload
$ systemctl --user enable pipewire.service
$ systemctl --user enable wireplumber.service
```
