{ pkgs, ... }:
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
      efi.canTouchEfiVariables = true;
      grub = {
        enable = true;
        device = "/dev/disk/by-id/usb-SanDisk_Extreme_55AE_323333364B31413030433539-0:0";
        efiSupport = true;
        configurationLimit = 40;
      };
    };
    initrd = {
      availableKernelModules = [ "nvme" "ahci" "xhci_pci" "usb_storage" "usbhid" "sd_mod" ];
      supportedFilesystems = [ "ext4" ];
      kernelModules = [ "amdgpu" "ext4" "nvme" "ahci" "xhci_pci" "usb_storage" "usbhid" "sd_mod" ];
    };
  };

  time.timeZone = "Europe/Warsaw";

  i18n.defaultLocale = "en_US.UTF-8";
  console = {
    font = "Lat2-Terminus16";
    keyMap = "us";
  };

  services.xserver.videoDrivers = [ "amdgpu" ];

  security.rtkit.enable = true;
  hardware.bluetooth.enable = true;
  hardware.bluetooth.powerOnBoot = true;
}
