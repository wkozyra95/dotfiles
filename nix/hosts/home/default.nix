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
      env = "home";
    })
    ./filesystems.nix
    ./system.nix
    ./boot.nix
    ../../nix-modules/common.nix
    ../../nix-modules/sway.nix
    ../../nix-modules/docker.nix
    ../../nix-modules/podman.nix
    ../../nix-modules/steam.nix
    ../../nix-modules/vm.nix
    ../../nix-modules/android.nix
    ../../nix-modules/printer.nix
    ({ config, lib, pkgs, ... }: {
      nixpkgs.overlays = overlays;
      home-manager = {
        extraSpecialArgs = {
          inherit custom;
        };
        useGlobalPkgs = true;
        useUserPackages = true;
        users.${config.myconfig.username} = (
          import ./home.nix config.myconfig.hm-modules
        );
      };

      environment.systemPackages = [
        pkgs.usbutils
        pkgs.tailscale
      ];
      services.tailscale.enable = true;
      systemd.oomd = {
        enable = true;
        enableUserSlices = true;
      };

      programs.nix-ld.enable = true;
      programs.nix-ld.libraries = with pkgs; [ ];
      programs.wireshark.enable = true;
      programs.wireshark.package = pkgs.wireshark;
      users.users.${config.myconfig.username} = {
        extraGroups = [ "wireshark" ];
      };
    })
  ];
}
