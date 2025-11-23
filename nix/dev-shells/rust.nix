{ pkgs, ... }:

pkgs.mkShell {
  packages = with pkgs; [
    openssl
    pkg-config
    rustfmt
    clippy
    cargo
    cargo-watch
    rust-analyzer
    rustc
  ];
}
