{ pkgs, lib, ... }:
let
  ffmpeg = pkgs.ffmpeg_7-full.override {
    withRtmp = false;
  };
in
pkgs.mkShell {
  env.LD_LIBRARY_PATH = lib.makeLibraryPath (with pkgs; [
    libx11
    libxext
    libxrandr
    libxfixes
    libxi
    libxcursor
    libxcomposite
    libxscrnsaver
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
