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
    home-manager
  ];

  networking.hostName = "wojtek-nix";
  networking.networkmanager.enable = true;

  networking.useDHCP = lib.mkDefault true;

  nixpkgs.hostPlatform = lib.mkDefault "x86_64-linux";
  powerManagement.cpuFreqGovernor = lib.mkDefault "powersave";
  hardware.cpu.amd.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;
  hardware.cpu.intel.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;

  hardware.amdgpu.initrd.enable = true;
  hardware.graphics = {
    enable = lib.mkForce true;
    enable32Bit = true;
  };

  # required to make wgpu project work without amdvlk installed
  # environment.variables.AMD_VULKAN_ICD = "RADV";
  # environment.variables.VK_ICD_FILENAMES = "/run/opengl-driver/share/vulkan/icd.d/radeon_icd.x86_64.json";

  system.stateVersion = "23.11";
}
