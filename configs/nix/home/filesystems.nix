{ config, pkgs, lib, modulesPath, ...}:
{
  imports =
    [ "${modulesPath}/installer/scan/not-detected.nix" ];

  fileSystems."/" =
    { device = "/dev/disk/by-label/BTRFS_ROOT";
      fsType = "btrfs";
      options = [
        "nodev" "ssd" "discard"
        "noatime" "space_cache=v2"
        "subvol=__current/root"
      ];
    };

  fileSystems."/home" =
    { device = "/dev/disk/by-label/BTRFS_ROOT";
      fsType = "btrfs";
      options = [ 
        "nosuid" "nodev" "ssd" "discard"
        "noatime" "space_cache=v2"
        "subvol=__current/home" 
      ];
    };

  fileSystems."/home/wojtek/.steam_volume" =
    { device = "/dev/disk/by-label/LOCAL_DATA_SSD";
      fsType = "btrfs";
      options = [ 
        "nosuid" "nodev" "ssd" "discard"
        "noatime" "space_cache=v2"
        "nofail"
        "subvol=steam"
      ];
      encrypted = {
        enable = true;
        blkDev = "/dev/disk/by-uuid/010f5771-8ef1-47a7-9237-a5da9bbb507b";
        keyFile = "/mnt-root/root/cryptlvm.keyfile";
        label = "local_data_ssd";
      };
      depends = [ "/home" ];
    };

  fileSystems."/home/wojtek/Videos" =
    { device = "/dev/disk/by-label/LOCAL_DATA_SSD";
      fsType = "btrfs";
      options = [ 
        "nosuid" "nodev" "ssd" "discard"
        "noatime" "space_cache=v2"
        "nofail"
        "subvol=video"
      ];
      encrypted = {
        enable = true;
        blkDev = "/dev/disk/by-uuid/010f5771-8ef1-47a7-9237-a5da9bbb507b";
        keyFile = "/mnt-root/root/cryptlvm.keyfile";
        label = "local_data_ssd";
      };
      depends = [ "/home" ];
    };

  fileSystems."/boot/efi" =
    { device = "/dev/disk/by-label/EFIBOOT";
      fsType = "vfat";
    };
}
