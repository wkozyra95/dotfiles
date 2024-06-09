{ pkgs, lib, ... }:
let
  ffmpeg =
    (if pkgs.stdenv.isDarwin then
      (pkgs.ffmpeg_6-full.override {
        x264 = pkgs.x264.overrideAttrs (old: {
          postPatch = old.postPatch + ''
            substituteInPlace Makefile --replace '$(if $(STRIP), $(STRIP) -x $@)' '$(if $(STRIP), $(STRIP) -S $@)'
          '';
        });
      })
    else
      pkgs.ffmpeg_6-full
    );
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
  ]);
  packages = with pkgs; [
    ffmpeg
    elixir
    nodejs_18
    rustfmt
    clippy
    rust-analyzer
    rustc
    cargo
  ];
}
