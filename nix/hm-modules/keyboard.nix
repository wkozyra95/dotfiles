{ pkgs, ... }:
let
  desktopItem = pkgs.makeDesktopItem {
    name = "keyboard-preview";
    desktopName = "Keyboard Layout";
    comment = "Show the keymap of the keyboard";
    exec = "mycli keyboard preview";
    terminal = false;
    icon = "input-keyboard";
    categories = [ "Utility" ];
  };
in
{
  # Viewer of `mycli keyboard preview`: the one that can show an image on a
  # transparent window. It draws in software, so the nix build also works
  # outside NixOS. Without it the preview opens with xdg-open.
  home.packages = [ pkgs.swayimg ];

  # Launcher entry ($mod+p). Linked into ~/.local/share instead of the nix
  # profile: outside NixOS (work-arch) sway starts without the profile in
  # XDG_DATA_DIRS, so fuzzel would not find it there.
  xdg.dataFile."applications/keyboard-preview.desktop".source =
    "${desktopItem}/share/applications/keyboard-preview.desktop";
}
