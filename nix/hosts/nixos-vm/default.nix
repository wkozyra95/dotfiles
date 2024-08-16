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
  system = "x86_64-linux";

  specialArgs = { inherit custom; };

  modules = [
    inputs.home-manager.nixosModules.home-manager
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
        extraSpecialArgs = { inherit custom; };
        useGlobalPkgs = true;
        useUserPackages = true;
        users.${config.myconfig.username} = (
          import ./home.nix config.myconfig.hm-modules
        );
      };
    })
  ];
}
