{ pkgs, lib, ... }:
# Notification daemon: notifications reach it over the session bus (notify-send,
# mycli, hostd). Rules give categories used by utils/notify their own look.
let
  # Rules apply in file order and the urgency_* sections count as rules, so the
  # category rules are written after them. Needs home-manager 26.05 (ordered
  # settings, https://github.com/nix-community/home-manager/issues/8961).
  afterUrgency = lib.hm.dag.entryAfter [ "urgency_low" "urgency_normal" "urgency_critical" "global" ];
in
{
  services.dunst = {
    enable = true;
    settings = {
      global = {
        origin = "top-right";
        offset = "(20, 20)";
        width = 400;
        # Grows with the content up to the limit
        height = "(0, 300)";
        font = "Source Code Pro 10";
        frame_width = 2;
        corner_radius = 8;
        gap_size = 8;
        padding = 12;
        horizontal_padding = 16;
        # No "(AU)" prefix marking actions and URLs
        show_indicators = false;
        # Menu for the actions of a notification, opened with the context binding
        dmenu = "${pkgs.fuzzel}/bin/fuzzel --dmenu";
        # Run with a URL picked from the menu. Copies instead of opening (default is /usr/bin/xdg-open).
        browser = "${pkgs.wl-clipboard}/bin/wl-copy";
        # Left runs the "default" action, middle shows the menu, right dismisses
        mouse_left_click = "do_action, close_current";
        mouse_right_click = "close_current";
      };
      # Colors per urgency (picked by the sender with -u). Timeout in seconds, 0 stays until dismissed.
      urgency_low = {
        background = "#262626";
        foreground = "#9a9a9a";
        frame_color = "#3a3a3a";
        timeout = 5;
      };
      urgency_normal = {
        background = "#2b3040";
        foreground = "#eeeeee";
        frame_color = "#4a90d9";
        timeout = 10;
      };
      urgency_critical = {
        background = "#4a1c1c";
        foreground = "#ffffff";
        frame_color = "#d94a4a";
        timeout = 0;
      };
      # Rule: `notify-send -c warning ...` gets a yellow frame regardless of urgency
      warning = afterUrgency {
        category = "warning";
        background = "#3d3418";
        foreground = "#ffffff";
        frame_color = "#d9b34a";
        timeout = 15;
      };
      # Rule: `notify-send -c error ...` looks like critical and stays until dismissed
      error = afterUrgency {
        category = "error";
        background = "#4a1c1c";
        foreground = "#ffffff";
        frame_color = "#d94a4a";
        timeout = 0;
      };
      # Rule: `notify-send -c success ...` confirms that something worked
      success = afterUrgency {
        category = "success";
        background = "#1c3a26";
        foreground = "#ffffff";
        frame_color = "#30a46c";
        timeout = 5;
      };
    };
  };
}
