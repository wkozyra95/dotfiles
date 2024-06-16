{ pkgs, config, ... }:
{
  nixpkgs.config.allowUnfree = true;
  hardware.enableAllFirmware = true;
  hardware.system76.kernel-modules.enable = true;
  boot = {
    kernelPackages = pkgs.linuxPackages_latest;
    kernelModules = [ "kvm-amd" "kvm-intel" ];
    extraModulePackages = [ ];
    supportedFilesystems = [ "ext4" ];
    loader = {
      efi.efiSysMountPoint = "/boot/efi";
      grub = {
        enable = true;
        device = "/dev/disk/by-id/usb-SanDisk_Extreme_55AE_323333364B31413030433539-0:0";
        efiSupport = true;
        efiInstallAsRemovable = true;
        configurationLimit = 40;
      };
    };
    initrd = {
      availableKernelModules = [
        "amdgpu"
        "nvidia"
        "ext4"
        "nvme"
        "ahci"
        "xhci_pci"
        "usbhid"
        "sd_mod"
        "uhci_hcd"
        "ehci_hcd"
        "ohci_hcd"
        "uas"
      ];
      supportedFilesystems = [ "ext4" ];
      kernelModules = [ ];
    };
  };

  time.timeZone = "Europe/Warsaw";

  i18n.defaultLocale = "en_US.UTF-8";
  console = {
    font = "Lat2-Terminus16";
    keyMap = "us";
  };

  services.xserver.videoDrivers = [ "amdgpu" "nvidia" ];

  hardware.nvidia = {
    modesetting.enable = true;
    powerManagement.enable = false;
    powerManagement.finegrained = false;
    open = false;
    nvidiaSettings = true;
    package = config.boot.kernelPackages.nvidiaPackages.stable;
  };

  security.rtkit.enable = true;
  hardware.bluetooth.enable = true;
  hardware.bluetooth.powerOnBoot = true;
}
