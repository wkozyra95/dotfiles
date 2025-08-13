{ nixpkgs, overlays, inputs }:

let
  system = "x86_64-linux";
  custom = {
    unstable = import inputs.nixpkgs-unstable {
      inherit system;
      config = { allowUnfree = true; };
    };
    neovim-nightly = inputs.neovim-nightly-overlay.packages.${system}.default;
  };
in

nixpkgs.lib.nixosSystem {
  inherit system;

  specialArgs = { inherit custom; };

  modules = [
    inputs.home-manager.nixosModules.home-manager
    (import ../../nix-modules/myconfig.nix {
      username = "wojtek";
      email = "wkozyra95@gmail.com";
      env = "usbdrive";
    })
    ./filesystems.nix
    ./system.nix
    ./boot.nix
    ../../nix-modules/common.nix
    ../../nix-modules/sway.nix
    ({ config, lib, pkgs, ... }: {
      nixpkgs.overlays = overlays;
      home-manager = {
        extraSpecialArgs = { inherit custom; };
        useGlobalPkgs = true;
        useUserPackages = true;
        backupFileExtension = "backup";
        users.${config.myconfig.username} = (
          import ./home.nix config.myconfig.hm-modules
        );
      };

      networking.firewall.allowedTCPPorts= [ 8002 ];
      networking.firewall.enable = false;
     # hardware.decklink.enable= true;
      environment.systemPackages = [
        pkgs.usbutils
        pkgs.pciutils
        pkgs.ffmpeg
        pkgs.vlc
        pkgs.mpv
        pkgs.cryptsetup
        pkgs.gptfdisk
        pkgs.btrfs-progs
        pkgs.parted
      ];
    })
  ];
}
