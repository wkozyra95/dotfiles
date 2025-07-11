{ pkgs, lib, ... }:
let
  ffmpeg = pkgs.ffmpeg_7-full.override {
    withRtmp = false;
  };
in
pkgs.mkShell {
  env.LD_LIBRARY_PATH = lib.makeLibraryPath (with pkgs; [
    xorg.libX11
    xorg.libXext
    xorg.libXrandr
    xorg.libXfixes
    xorg.libXi
    xorg.libXcursor
    xorg.libXcomposite
    xorg.libXScrnSaver
    alsa-lib
    openssl
    ffmpeg
  ]);
  packages = with pkgs; [
    ffmpeg
    elixir
    nodejs
    rustfmt
    clippy
    rust-analyzer
    rustc
    cargo
  ];
}
