{ pkgs, ... }:
{
  nixpkgs.config.allowUnfree = true;
  hardware.enableAllFirmware = true;
  boot = {
    kernelPackages = pkgs.linuxPackages_latest;
    kernelModules = [
      "kvm-amd"
    ];
    extraModulePackages = [ ];
    supportedFilesystems = [ "btrfs" "ext4" "vfat" ];
    loader = {
      efi.efiSysMountPoint = "/efi";
      efi.canTouchEfiVariables = true;
      grub = {
        enable = true;
        # To skip grub installation
        # device = "nodev";
        device = "/dev/disk/by-id/nvme-KINGSTON_SA2000M81000G_50026B768404F6F5";
        efiSupport = true;
        configurationLimit = 40;
        theme = ../../../configs/grub-theme;
        splashImage = ../../../configs/grub-theme/background.png;
      };
    };
    initrd = {
      availableKernelModules = [
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
        "usb_storage"
      ];
      kernelModules = [
        "amdgpu"
        "uas"
        "usbcore"
        "usb_storage"
        "vfat"
        "nls_cp437"
        "nls_iso8859_1"
      ];
      # Mount USB key before trying to decrypt root filesystem
      postDeviceCommands = pkgs.lib.mkBefore ''
        mkdir -m 0755 -p /usb-secrets
        sleep 2 # To make sure the usb key has been loaded
        mount -n -t vfat -o ro `findfs LABEL=SECRETS_USB` /usb-secrets
      '';
      luks.devices = {
        root = {
          device = "/dev/disk/by-uuid/c55ce2d0-79c4-4e3f-af69-b9f41ee31246";
          keyFile = "/usb-secrets/cryptlvm.keyfile";
          fallbackToPassword = true;
          preLVM = false;
        };
      };

    };
    kernel.sysctl = {
      "net.ipv4.ip_forward" = 1;
      "net.ipv6.conf.all.forwarding" = 1;
    };
  };

  time.timeZone = "Europe/Warsaw";

  i18n.defaultLocale = "en_US.UTF-8";
  console = {
    font = "Lat2-Terminus16";
    keyMap = "us";
  };

  services.xserver.videoDrivers = [ "amdgpu" ];

  services.openssh.enable = true;
  security.rtkit.enable = true;
  hardware.bluetooth.enable = true;
  hardware.bluetooth.powerOnBoot = true;
}
