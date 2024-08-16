{ overlays, nixpkgs, inputs }:

let
  system = "x86_64-linux";
  pkgs = import nixpkgs {
    inherit system;
    config = { allowUnfree = true; };
  };
  custom = {
    unstable = import inputs.nixpkgs-unstable {
      inherit system;
      config = { allowUnfree = true; };
    };
    neovim-nightly = inputs.neovim-nightly-overlay.packages.${system}.default;
  };
in

inputs.home-manager.lib.homeManagerConfiguration {
  pkgs = pkgs;

  extraSpecialArgs = { inherit custom; };

  modules = [
    (import ../../hm-modules/myconfig.nix {
      username = "wojtek";
      email = "wojciech.kozyra@swmansion.com";
      env = "work";
    })
    ../../common.nix
    ../../hm-modules/common.nix
    ../../hm-modules/git.nix
    ../../hm-modules/vim.nix
    ../../hm-modules/neovim.nix
    ../../hm-modules/dotfiles.nix
    ({ config, lib, pkgs, ... }: {
      home.username = config.myconfig.username;
      home.homeDirectory = "/home/${config.myconfig.username}";

      nixpkgs.overlays = overlays;
      nix.package = pkgs.nix;

      myconfig = {
        git.signingKey = "C577 851A 36BB 9BBA CF11  8A44 CA11 EA63 4382 0983";
      };

      programs.gpg.enable = true;
      services.gpg-agent = {
        enable = true;
        pinentryPackage = pkgs.pinentry-curses;
        enableSshSupport = true;
        enableExtraSocket = true;
      };

      programs.home-manager.enable = true;
      home.stateVersion = "23.11";
    })
  ];
}
