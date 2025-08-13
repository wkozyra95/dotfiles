{ pkgs, ... }:
{
  nixpkgs.config.allowUnfree = true;
  hardware.enableAllFirmware = true;
  boot = {
    kernelPackages = pkgs.linuxPackages_latest;
    kernelModules = [ "kvm-amd" ];
    extraModulePackages = [ ];
    supportedFilesystems = [ "btrfs" ];
    loader = {
      efi.efiSysMountPoint = "/boot/efi";
      efi.canTouchEfiVariables = true;
      grub = {
        enable = true;
        # To skip grub installation
        # device = "nodev";
        device = "/dev/disk/by-id/nvme-Samsung_SSD_990_EVO_Plus_4TB_S7U9NJ0Y508453F";
        efiSupport = true;
        enableCryptodisk = true;
        configurationLimit = 40;
        theme = ../../../configs/grub-theme;
        splashImage = ../../../configs/grub-theme/background.png;
      };
    };
    initrd = {
      availableKernelModules = [ "nvme" "ahci" "xhci_pci" "usb_storage" "usbhid" "sd_mod" ];
      kernelModules = [ "amdgpu" ];
      luks.devices = {
        root = {
          device = "/dev/disk/by-uuid/6c560c44-363f-4e19-91a1-7709bbc0d9b6";
          keyFile = "/root/cryptlvm.keyfile";
          fallbackToPassword = true;
        };
      };
      secrets = {
        "/root/cryptlvm.keyfile" = "/root/cryptlvm.keyfile";
      };
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
