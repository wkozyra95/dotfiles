{ home-manager, overlays, pkgs }:

home-manager.lib.homeManagerConfiguration {
  pkgs = pkgs;
  modules = [
    (import ../hm-modules/myconfig.nix {
      username = "wojtek";
      email = "wojciechkozyra@swmansion.com";
      env = "work";
    })
    ../common.nix
    ../hm-modules/common.nix
    ../hm-modules/git.nix
    ../hm-modules/vim.nix
    ../hm-modules/neovim.nix
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
