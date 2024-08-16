{
  description = "Generic dev shell for Rust project";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];
      perSystem = { config, self', inputs', pkgs, system, lib, ... }:
        {
          devShells = {
            default = pkgs.mkShell {
              packages = with pkgs; [
                rustfmt
                clippy
                cargo
                cargo-watch
                rust-analyzer
                rustc
              ];
            };
          };
        };
    };
}
