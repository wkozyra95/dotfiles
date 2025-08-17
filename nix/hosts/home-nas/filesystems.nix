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
      neededForBoot = true;
    };

  fileSystems."/efi" =
    {
      device = "/dev/disk/by-label/EFI";
      fsType = "vfat";
      depends = [ "/" ];
    };

  fileSystems."/boot" =
    {
      device = "/dev/disk/by-label/BOOT";
      fsType = "ext4";
      depends = [ "/" ];
    };
}
