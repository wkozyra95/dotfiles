{ nix-darwin, overlays, inputs }:

let
  system = "aarch64-darwin";
  custom = {
    unstable = import inputs.nixpkgs-unstable {
      inherit system;
      config = { allowUnfree = true; };
    };
    neovim-nightly = inputs.neovim-nightly-overlay.packages.${system}.default;
  };
in

nix-darwin.lib.darwinSystem {
  inherit system;

  specialArgs = { inherit custom; };

  modules = [
    inputs.home-manager.darwinModules.home-manager
    (import ../../nix-modules/myconfig.nix {
      username = "wojciechkozyra";
      email = "wojciech.kozyra@swmansion.com";
      env = "macbook";
    })
    (import ../../nix-modules/common.nix)
    ({ config, lib, pkgs, ... }: {
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

      time.timeZone = "Europe/Warsaw";

      networking.hostName = "wojtek-mac";

      nixpkgs = {
        hostPlatform = lib.mkDefault "aarch64-darwin";
        overlays = overlays;
      };

      services.nix-daemon.enable = true;
      users.users.${config.myconfig.username} = {
        home = "/Users/${config.myconfig.username}";
      };

      system.stateVersion = 4;
    })
  ];
}
