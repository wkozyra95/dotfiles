{ inputs, customPackages }:

inputs.nixpkgs.lib.nixosSystem {
  system = "x86_64-linux";

  modules = [
    (import ./filesystems.nix)
    (import ./users.nix)
    (import ./boot.nix)
    (import ../common.nix inputs)
    (import ../modules/sway.nix)
    (import ../modules/languages.nix)
    (import ../modules/docker.nix)
    (import ../modules/steam.nix "wojtek")
    (import ../modules/neovim.nix inputs.neovim-nightly-overlay.overlay)
    ({ config, lib, pkgs, ... }: {
      environment.systemPackages = with pkgs; [ amdvlk ] ++ customPackages;

      networking.hostName = "wojtek-nix";
      networking.networkmanager.enable = true;

      networking.useDHCP = lib.mkDefault true;
      networking.interfaces.enp39s0.useDHCP = lib.mkDefault true;
      networking.interfaces.wlp41s0.useDHCP = lib.mkDefault true;

      nixpkgs.hostPlatform = lib.mkDefault "x86_64-linux";
      powerManagement.cpuFreqGovernor = lib.mkDefault "powersave";
      hardware.cpu.amd.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;

      hardware.opengl = {
        enable = true;
        driSupport = true;
        driSupport32Bit = true;
      };

      system.stateVersion = "23.11";
    })
  ];
}
