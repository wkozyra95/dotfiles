{ pkgs, ... }:
{
  environment.systemPackages = with pkgs; [
    # go
    go
    gopls
    golines
    gofumpt
    golangci-lint-langserver
    golangci-lint

    # nix
    nil

    # lua
    sumneko-lua-language-server

    # typescipt
    nodePackages.typescript-language-server
    vscode-langservers-extracted

    # rust
    clippy
    rust-analyzer

    # c/c++
    gcc

    # other
    efm-langserver
  ];
}
