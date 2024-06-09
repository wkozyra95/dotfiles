{ pkgs, ... }:
{
  nixpkgs.config.allowUnfree = true;
  hardware.enableAllFirmware = true;
  hardware.system76.kernel-modules.enable = true;
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
        device = "/dev/disk/by-id/nvme-KINGSTON_SA2000M81000G_50026B768404F6F5";
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
          device = "/dev/disk/by-uuid/546d1104-c026-4072-9e46-1b52fed323a5";
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
