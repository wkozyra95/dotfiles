{ pkgs, ... }:
let
  dbus-sway-environment = pkgs.writeTextFile {
    name = "dbus-sway-environment";
    destination = "/bin/dbus-sway-environment";
    executable = true;

    text = ''
      dbus-update-activation-environment --systemd WAYLAND_DISPLAY XDG_CURRENT_DESKTOP=sway
      systemctl --user stop pipewire pipewire-media-session xdg-desktop-portal xdg-desktop-portal-wlr
      systemctl --user start pipewire pipewire-media-session xdg-desktop-portal xdg-desktop-portal-wlr
    '';
  };
in
{
  fonts.packages = with pkgs; [
    source-code-pro
    nerdfonts
    corefonts
  ];
  environment.systemPackages = with pkgs; [
    firefox
    dbus
    dbus-sway-environment
    j4-dmenu-desktop
    alacritty
    bemenu
    diff-so-fancy
    pavucontrol
    grim
    slurp
    wf-recorder
    swaylock
    btop
    wl-clipboard
    playerctl
    pamixer
    libnotify
    dunst # notification daemon
  ];
  programs.sway.enable = true;
  programs.xwayland.enable = true;

  security.rtkit.enable = true;
  services.pipewire = {
    enable = true;
    alsa.enable = true;
    alsa.support32Bit = true;
    pulse.enable = true;
    #jack.enable = true;
  };

  services.dbus.enable = true;
  xdg.portal = {
    enable = true;
    wlr.enable = true;
    # gtk portal needed to make gtk apps happy
    extraPortals = [ pkgs.xdg-desktop-portal-gtk ];
  };
}
