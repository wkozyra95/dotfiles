{ nixpkgs, home-manager, overlays }:

nixpkgs.lib.nixosSystem {
  system = "x86_64-linux";

  modules = [
    home-manager.nixosModules.home-manager
    (import ../nix-modules/myconfig.nix {
      username = "wojtek";
      email = "wkozyra95@gmail.com";
    })
    ./filesystems.nix
    ./system.nix
    ./boot.nix
    ../common.nix
    ../nix-modules/sway.nix
    ../nix-modules/docker.nix
    ../nix-modules/steam.nix
    ../nix-modules/vm.nix
    ({ config, lib, pkgs, ... }: {
      nixpkgs.overlays = overlays;
      home-manager = {
        useGlobalPkgs = true;
        useUserPackages = true;
        users.${config.myconfig.username} = (
          import ./home.nix config.myconfig.hm-modules
        );
      };
    })
  ];
}
