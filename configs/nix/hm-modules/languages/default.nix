{ pkgs, ... }:
{
  imports = [
    ./build-tools.nix
    ./go.nix
    ./rust.nix
  ];
  config = {
    home.packages = with pkgs; [
      nodejs_18

      # Language servers

      ## nix
      nil

      ## lua
      sumneko-lua-language-server

      ## typescipt
      nodePackages.typescript-language-server
      vscode-langservers-extracted

      ## other
      efm-langserver
    ];
  };
}
