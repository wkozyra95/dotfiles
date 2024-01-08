{ pkgs, config, ... }:
let
  dotfilesSymlink = path:
    config.lib.file.mkOutOfStoreSymlink "${config.home.homeDirectory}/.dotfiles/${path}";
in
{
  imports = [
    ../common.nix
    ../hm-modules/common.nix
  ];

  config = {
    home.file = {
      ".gitconfig".source = dotfilesSymlink "env/macbook/gitconfig";
      ".gitignore".source = dotfilesSymlink "env/macbook/gitignore";
    };

    home.packages = with pkgs; [
      nodejs_18
      nil
      sumneko-lua-language-server
      nodePackages.typescript-language-server
      vscode-langservers-extracted
      efm-langserver
    ];

    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
