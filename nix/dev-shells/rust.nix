{ pkgs, ... }:

pkgs.mkShell {
  packages = with pkgs; [
    rustfmt
    clippy
    cargo
    cargo-watch
    rust-analyzer
    rustc
  ];
}
