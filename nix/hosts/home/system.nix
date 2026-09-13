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
    sshfs
  ];

  # Puts mount.sshfs where `mount -t sshfs` looks for it.
  system.fsPackages = [ pkgs.sshfs ];
  # Required for allow_other below: systemd mounts as root, but the mount has
  # to be usable by wojtek.
  programs.fuse.userAllowOther = true;

  # The NAS's /storage as a local path. Nothing connects at boot — systemd only
  # watches the mountpoint and mounts on first access, then unmounts once idle.
  # The short timeouts are deliberate: the NAS is often powered off, and a dead
  # host should fail in seconds rather than hang whatever touched the path.
  fileSystems."/mnt/nas" = {
    device = "wojtek@192.168.100.5:/storage";
    fsType = "sshfs";
    # There is nothing to fsck over ssh.
    noCheck = true;
    options = [
      "noauto"
      "x-systemd.automount"
      "_netdev"
      "x-systemd.idle-timeout=60"
      "x-systemd.mount-timeout=10s"
      "allow_other"
      # Fail fast when the NAS is down. No `reconnect` — retrying forever is
      # right for a flaky link, wrong for a machine that is deliberately off.
      "ConnectTimeout=5"
      "ServerAliveInterval=5"
      "ServerAliveCountMax=2"
      # The mount runs as root, which has neither the key nor the host key, so
      # both have to be pointed at wojtek's. id_rsa is the one the NAS accepts.
      "IdentityFile=/home/wojtek/.ssh/id_rsa"
      "UserKnownHostsFile=/home/wojtek/.ssh/known_hosts"
      "StrictHostKeyChecking=accept-new"
    ];
  };

  networking.hostName = "wojtek-nix";
  networking.networkmanager.enable = true;
  networking.networkmanager.insertNameservers = [ "8.8.8.8" ];

  # tailscaled registers the MagicDNS suffix with resolved as a routing domain,
  # so only *.ts.net goes to 100.100.100.100 and the rest still goes to 8.8.8.8.
  # Listing 100.100.100.100 in insertNameservers instead does not work: glibc
  # asks 8.8.8.8 first, takes its NXDOMAIN as authoritative, and never falls
  # through to the second server.
  services.resolved.enable = true;

  networking.useDHCP = lib.mkDefault true;
  networking.interfaces.eno1.useDHCP = lib.mkDefault true;

  nixpkgs.hostPlatform = lib.mkDefault "x86_64-linux";
  powerManagement.cpuFreqGovernor = lib.mkDefault "powersave";
  hardware.cpu.amd.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;

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
