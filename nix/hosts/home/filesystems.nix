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

  fileSystems."/home/wojtek/Videos" =
    {
      device = "/dev/disk/by-label/LOCAL_DATA_HDD";
      fsType = "btrfs";
      options = [
        "nosuid"
        "nodev"
        "discard"
        "noatime"
        "space_cache=v2"
        "nofail"
        "subvol=video"
      ];
      encrypted = {
        enable = true;
        blkDev = "/dev/disk/by-uuid/a2f19d9b-359f-470f-a5f6-71faf8dc6d8e";
        keyFile = "/mnt-root/root/cryptlvm.keyfile";
        label = "local_data_hdd";
      };
      depends = [ "/home" ];
    };

  fileSystems."/boot/efi" =
    {
      device = "/dev/disk/by-label/EFIBOOT";
      fsType = "vfat";
    };
}
