{ modulesPath, ... }:
{
  imports =
    [ "${modulesPath}/installer/scan/not-detected.nix" ];

  fileSystems."/" =
    {
      device = "/dev/disk/by-label/BTRFS_ROOT";
      fsType = "btrfs";
      options = [
        "nodev"
        "ssd"
        "discard"
        "noatime"
        "space_cache=v2"
        "subvol=root"
      ];
    };

  fileSystems."/boot/efi" =
    {
      device = "/dev/disk/by-label/EFIBOOT";
      fsType = "vfat";
    };
}
