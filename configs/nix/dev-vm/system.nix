{ pkgs, config, lib, ... }:

{
  users = {
    mutableUsers = true;
    users = {
      ${config.myconfig.username} = {
        isNormalUser = true;
        home = "/home/${config.myconfig.username}";
        extraGroups = [ "wheel" ];
        shell = pkgs.zsh;
        initialPassword = "wojtek";
      };

      root = {
        home = "/root";
      };
    };
  };

  virtualisation.vmVariant = {
    virtualisation.qemu.options = [
      "-vga none"
      "-device virtio-gpu-pci"
      "-display gtk,grab-on-hover=on"
    ];
  };

  networking.hostName = "dev";
  networking.networkmanager.enable = true;
  networking.useDHCP = lib.mkDefault true;

  programs.zsh.enable = true;

  nixpkgs.config.allowUnfree = true;

  hardware.opengl = {
    enable = lib.mkForce true;
    driSupport = true;
    driSupport32Bit = true;
    extraPackages = [ ];
  };

  system.stateVersion = "23.11";
}
