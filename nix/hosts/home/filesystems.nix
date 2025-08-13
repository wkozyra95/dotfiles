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
        "subvol=__current/root"
      ];
    };

  fileSystems."/home" =
    {
      device = "/dev/disk/by-label/BTRFS_ROOT";
      fsType = "btrfs";
      options = [
        "nosuid"
        "nodev"
        "ssd"
        "discard"
        "noatime"
        "space_cache=v2"
        "subvol=__current/home"
      ];
    };

  fileSystems."/boot/efi" =
    {
      device = "/dev/disk/by-label/EFIBOOT";
      fsType = "vfat";
    };
}
