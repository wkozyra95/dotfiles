{ nixpkgs, nixpkgs-unstable, home-manager, overlays }:

let
  system = "x86_64-linux";
  unstable = import nixpkgs-unstable {
    inherit system;
    config = { allowUnfree = true; };
  };
in

nixpkgs.lib.nixosSystem {
  inherit system;

  specialArgs = { inherit unstable; };

  modules = [
    home-manager.nixosModules.home-manager
    (import ../../nix-modules/myconfig.nix {
      username = "wojtek";
      email = "wkozyra95@gmail.com";
      env = "home";
    })
    ./filesystems.nix
    ./system.nix
    ./boot.nix
    ../../common.nix
    ../../nix-modules/sway.nix
    ../../nix-modules/docker.nix
    ../../nix-modules/steam.nix
    ../../nix-modules/vm.nix
    ../../nix-modules/android.nix
    ../../nix-modules/printer.nix
    ({ config, lib, pkgs, ... }: {
      nixpkgs.overlays = overlays;
      home-manager = {
        extraSpecialArgs = {
          inherit unstable;
        };
        useGlobalPkgs = true;
        useUserPackages = true;
        users.${config.myconfig.username} = (
          import ./home.nix config.myconfig.hm-modules
        );
      };

      environment.systemPackages = [
        pkgs.usbutils
      ];

      users.users.${config.myconfig.username} = {
        extraGroups = [ "wireshark" ];
      };
    })
  ];
}
