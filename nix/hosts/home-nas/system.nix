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
        openssh.authorizedKeys.keys = [
          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCsauS5jVzD5Wc+Ekhs/IgxFIXS/P+4JtluuSly5h5o0+b1cXi4q0c2Z9o7u0lp6bticWX2IS+1XVzZsbbVNtPkENtstgG979lbHbWMs/dpoqgUicZzLvRgbG0NxF13cQBnQ2vafLlImvUhGIu0Prep4XRc6iH8QLmgUgG9glgZZxCAa4gWtwUA6wqMyLcYGuMjP6dHnuUP6XHfmMMG32p42UZ0Qu/IiEuphrwLPB/YWm/9kyLt/9gSW4fxd5jxDfF2Mbv4ifT9q2vJhLmgcwRosnNUAVVC69mF6lgGgJJwdSoHvtrfYPA4MJyfe5QeDgVpO118xopvYu4j74EBQ6MtUUnXi+IXct04I1s+3Bxe9h/hn1DwwGaLLfagvu97gRytCcVoMCPIfx4vYljc/Lz+7iNYp3wfRU6TSaUNnQL/ao0NaOrbIx6YQUcFKRT2kgpqiYTt4FENOeXsRyv2SqYmLRWmJA40KmIEPp4nDdnXhmUnaGNWz1KEZGiYWf0DXl8="
          "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEcWfDrnt0InlMBPf1bJDN9ZvkdakkkSSndpwDjfF4Y0"
        ];
      };

      root = {
        home = "/root";
      };
    };
  };

  # Allow pushing locally built (unsigned) closures with
  # `nixos-rebuild --target-host ${config.myconfig.username}@...`
  nix.settings.trusted-users = [ "root" config.myconfig.username ];

  programs.zsh.enable = true;

  environment.systemPackages = with pkgs; [
    vulkan-tools
    home-manager
  ];

  networking.hostName = "wojtek-nas";
  networking.networkmanager.enable = true;
  networking.networkmanager.insertNameservers = [ "8.8.8.8" ];

  networking.useDHCP = lib.mkDefault true;
  # Keep the NIC armed for Wake-on-LAN magic packets (needs "Resume By PCI-E Device" in BIOS).
  networking.interfaces.enp39s0.wakeOnLan.enable = true;
  #networking.interfaces.eno1.useDHCP = lib.mkDefault true;
  #networking.interfaces.wlp11s0.useDHCP = lib.mkDefault true;

  nixpkgs.hostPlatform = lib.mkDefault "x86_64-linux";
  # acpi-cpufreq + "powersave" pins the CPU at its lowest P-state (2.2 GHz) even under
  # load; schedutil idles at the same 2.2 GHz but boosts when there is work.
  powerManagement.cpuFreqGovernor = lib.mkDefault "schedutil";
  hardware.cpu.amd.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;

  hardware.amdgpu.initrd.enable = true;
  hardware.graphics = {
    enable = lib.mkForce true;
    enable32Bit = true;
  };

  system.stateVersion = "23.11";
}
