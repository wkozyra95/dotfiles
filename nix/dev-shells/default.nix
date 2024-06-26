{ pkgs, ... }@args:

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
      nixpkgs-fmt
    ];
  };
  devops = import ./devops.nix args;
  elixir = import ./elixir.nix args;
  membrane = import ./membrane.nix args;
  rust = import ./rust.nix args;
}
