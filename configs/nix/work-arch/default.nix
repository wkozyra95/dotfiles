{ home-manager, myConfigModule, overlays }:

home-manager.lib.homeManagerConfiguration {
  modules = [
    myConfigModule
    ../hm-modules/common.nix
    ../hm-modules/common-desktop.nix
    ../hm-modules/languages
    ({ config, lib, pkgs, ... }: {
      home.username = config.myconfig.username;
      home.homeDirectory = "/home/${config.myconfig.username}";

      nixpkgs.overlays = overlays;

      programs.home-manager.enable = true;
      home.stateVersion = "23.11";
    })
  ];
}
