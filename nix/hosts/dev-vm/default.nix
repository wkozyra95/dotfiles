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
      email = "wkozyra95@gmail.com";
      env = "dev-vm";
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

      programs.home-manager.enable = true;
      home.stateVersion = "23.11";
    })
  ];
}
