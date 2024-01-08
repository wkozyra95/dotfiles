{ home-manager, overlays, pkgs }:

home-manager.lib.homeManagerConfiguration {
  pkgs = pkgs;
  modules = [
    (import ../nix-modules/myconfig.nix {
      username = "wojtek";
      email = "wojciechkozyra@swmansion.com";
    })
    ../common.nix
    ../hm-modules/common.nix
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
          ".gitconfig".source = dotfilesSymlink "env/work/gitconfig";
          ".gitignore".source = dotfilesSymlink "env/work/gitignore";
        };

      home.packages = with pkgs; [
        nodejs_18
        nil
        sumneko-lua-language-server
        nodePackages.typescript-language-server
        vscode-langservers-extracted
        efm-langserver
      ];

      nixpkgs.overlays = overlays;
      nix.package = pkgs.nix;

      programs.home-manager.enable = true;
      home.stateVersion = "23.11";
    })
  ];
}
