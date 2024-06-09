{ pkgs, ... }:
{
  home.packages = with pkgs; [
    go
    gopls
    golines
    gofumpt
    golangci-lint-langserver
    golangci-lint
  ];
}
