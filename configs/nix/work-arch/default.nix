{ home-manager, overlays }:

home-manager.lib.homeManagerConfiguration {
  modules = [
    (import ../nix-modules/myconfig.nix {
      username = "wojtek";
      email = "wojciechkozyra@swmansion.com";
    })
    ../hm-modules/common.nix
    ../hm-modules/common-desktop.nix
    ../hm-modules/languages
    ({ config, lib, pkgs, ... }: {
      home.username = config.myconfig.username;
      home.homeDirectory = "/home/${config.myconfig.username}";

      home.file =
        let
          dotfilesSymlink = path:
            config.lib.file.mkOutOfStoreSymlink
              "${config.home.homeDirectory}/.dotfiles/${path}";
        in
        {
          ".gitconfig".source = dotfilesSymlink "env/home/gitconfig";
          ".gitignore".source = dotfilesSymlink "env/home/gitignore";
        };

      nixpkgs.overlays = overlays;

      programs.home-manager.enable = true;
      home.stateVersion = "23.11";
    })
  ];
}
