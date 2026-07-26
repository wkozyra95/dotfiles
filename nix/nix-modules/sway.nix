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
    nerd-fonts.droid-sans-mono
    corefonts
  ];
  environment.systemPackages = with pkgs; [
    dbus
    dbus-sway-environment
    j4-dmenu-desktop
    # alacritty
    bemenu
    fuzzel
    waybar
    pavucontrol
    grim
    slurp
    wf-recorder
    swaylock
    wl-clipboard
    playerctl
    pamixer
    libnotify
    rhythmbox
    mpv
    vlc
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
    jack.enable = true;
    wireplumber = {
      extraConfig =
        # pw-cli info all
        {
          "wh-1000xm3-ldac-hq" = {
            "monitor.bluez.rules" = [
              {
                matches = [
                  {
                    "device.name" = "~bluez_card.*";
                    "device.product.id" = "0x0cd3";
                    "device.vendor.id" = "usb:054c";
                  }
                ];
                actions = {
                  update-props = {
                    "bluez5.codecs" = [ "ldac" "aac" ];
                    "bluez5.a2dp.ldac.quality" = "hq";
                  };
                };
              }
            ];
          };
        };
    };
  };

  services.dbus.enable = true;
  # Unlock the gnome-keyring login keyring at TTY login (same password), so the
  # Secret Service is usable for stored creds (e.g. docker-credential-secretservice).
  security.pam.services.login.enableGnomeKeyring = true;
  xdg.portal = {
    enable = true;
    wlr.enable = true;
    # gtk portal needed to make gtk apps happy
    extraPortals = [ pkgs.xdg-desktop-portal-gtk ];
  };

  myconfig.hm-modules = [
    {
      # gnome-keyring provides the Secret Service (org.freedesktop.secrets) that
      # docker-credential-secretservice (installed in common.nix) talks to.
      # Limit to the "secrets" component so it doesn't hijack the ssh-agent.
      services.gnome-keyring = {
        enable = true;
        components = [ "secrets" ];
      };
      programs.alacritty.enable = true;
      gtk = {
        enable = true;
        gtk3.extraConfig = {
          gtk-application-prefer-dark-theme = true;
        };
        gtk4.extraConfig = {
          gtk-application-prefer-dark-theme = true;
        };
      };
    }
  ];
}
