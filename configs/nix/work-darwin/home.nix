{ config, ... }:
let
  dotfilesSymlink = path:
    config.lib.file.mkOutOfStoreSymlink "${config.home.homeDirectory}/.dotfiles/${path}";
in
{
  imports = [
    ../hm-modules/common.nix
    ../hm-modules/languages
  ];

  config = {
    home.file = {
      ".gitconfig".source = dotfilesSymlink "env/home/gitconfig";
      ".gitignore".source = dotfilesSymlink "env/home/gitignore";
    };

    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
