{ pkgs, ... }:

{
  default = pkgs.mkShell {
    packages = with pkgs; [
      go
      modd
      gopls
      golines
      gofumpt
      golangci-lint-langserver
      golangci-lint
    ];
  };
}
