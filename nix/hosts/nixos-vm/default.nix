{ nixpkgs, home-manager, overlays, nixpkgs-unstable }:

let
  system = "x86_64-linux";
  unstable = import nixpkgs-unstable {
    inherit system;
    config = { allowUnfree = true; };
  };
in

nixpkgs.lib.nixosSystem {
  system = "x86_64-linux";

  specialArgs = { inherit unstable; };

  modules = [
    home-manager.nixosModules.home-manager
    (import ../../nix-modules/myconfig.nix {
      username = "wojtek";
      email = "wkozyra95@gmail.com";
      env = "dev-vm";
    })
    ./system.nix
    ../../nix-modules/sway.nix
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
    })
  ];
}
