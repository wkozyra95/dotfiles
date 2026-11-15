{ pkgs, ... }:
{
  # Viewer of `mycli keyboard preview`: the one that can show an image on a
  # transparent window. It draws in software, so the nix build also works
  # outside NixOS. Without it the preview opens with xdg-open.
  home.packages = [ pkgs.swayimg ];

  # Launcher entry ($mod+p)
  xdg.desktopEntries.keyboard-preview = {
    name = "Keyboard Layout";
    comment = "Show the keymap of the keyboard";
    exec = "mycli keyboard preview";
    terminal = false;
    icon = "input-keyboard";
    categories = [ "Utility" ];
  };
}
