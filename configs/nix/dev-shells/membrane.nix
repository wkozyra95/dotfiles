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
  # https://github.com/NixOS/nixpkgs/blob/master/pkgs/development/libraries/libcef/default.nix#L33
  libcefDependencies = with pkgs;  [
    glib
    nss
    nspr
    atk
    at-spi2-atk
    expat
    xorg.libxcb
    libxkbcommon
    xorg.libX11
    xorg.libXcomposite
    xorg.libXdamage
    xorg.libXext
    xorg.libXfixes
    xorg.libXrandr
    mesa
    gtk3
    pango
    cairo
    dbus
    at-spi2-core
    cups
    xorg.libxshmfence
  ] ++ (
    pkgs.lib.optionals pkgs.stdenv.isLinux [
      libdrm
      alsa-lib
    ]
  );

  libs = with pkgs; [
    ffmpeg
    openssl
    libopus
    libGL
    mesa.drivers
    vulkan-loader
    mesa.drivers
    pkg-config
    llvmPackages_16.clang
    SDL2
  ];
in
pkgs.mkShell {
  env.LD_LIBRARY_PATH = lib.makeLibraryPath (libcefDependencies ++ libs);
  packages = with pkgs; [
    elixir
    nodejs_18
    rustfmt
    clippy
    rust-analyzer
  ] ++ libs ++ libcefDependencies;
  nativeBuildInputs = with pkgs; [
    elixir
    nodejs_18
    rustfmt
    clippy
    rust-analyzer
  ] ++ libs ++ libcefDependencies;
}
