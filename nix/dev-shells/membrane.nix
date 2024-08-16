{ pkgs, lib, ... }:
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
    ffmpeg_7-full
  ]);
  packages = with pkgs; [
    ffmpeg_7-full
    elixir
    nodejs_18
    rustfmt
    clippy
    rust-analyzer
    rustc
    cargo
  ];
}
