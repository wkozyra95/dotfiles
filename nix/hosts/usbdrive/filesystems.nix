{ modulesPath, ... }:
{
  imports =
    [ "${modulesPath}/installer/scan/not-detected.nix" ];

  fileSystems."/" =
    {
      device = "/dev/disk/by-label/USB_ROOT";
      fsType = "ext4";
    };

  fileSystems."/boot/efi" =
    {
      device = "/dev/disk/by-label/USB_EFI";
      fsType = "vfat";
    };
}
