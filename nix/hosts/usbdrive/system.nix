{ pkgs, config, lib, ... }:

{
  users = {
    mutableUsers = true;
    users = {
      ${config.myconfig.username} = {
        isNormalUser = true;
        home = "/home/${config.myconfig.username}";
        extraGroups = [ "wheel" "networkmanager" ];
        shell = pkgs.zsh;
        initialPassword = "password";
      };

      root = {
        home = "/root";
      };
    };
  };

  programs.zsh.enable = true;

  environment.systemPackages = with pkgs; [
    vulkan-tools
    amdvlk
    home-manager
  ];

  networking.hostName = "wojtek-nix";
  networking.networkmanager.enable = true;

  networking.useDHCP = lib.mkDefault true;
  networking.interfaces.enp39s0.useDHCP = lib.mkDefault true;
  networking.interfaces.wlp41s0.useDHCP = lib.mkDefault true;

  nixpkgs.hostPlatform = lib.mkDefault "x86_64-linux";
  powerManagement.cpuFreqGovernor = lib.mkDefault "powersave";
  hardware.cpu.amd.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;

  hardware.opengl = {
    enable = lib.mkForce true;
    driSupport = true;
    driSupport32Bit = true;
    extraPackages = [ pkgs.amdvlk ];
  };

  # required to make wgpu project work without amdvlk installed
  # environment.variables.AMD_VULKAN_ICD = "RADV";
  # environment.variables.VK_ICD_FILENAMES = "/run/opengl-driver/share/vulkan/icd.d/radeon_icd.x86_64.json";

  system.stateVersion = "23.11";
}
