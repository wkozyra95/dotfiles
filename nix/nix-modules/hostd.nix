{ lanInterface
, port ? 7420
  # directory for file transfers, /files endpoints are disabled without it
, filesDir ? null
  # symlink to filesDir, e.g. in the home directory
, filesLink ? null
  # Run as this user instead of the hostd system user and enable /notify: desktop
  # notifications need the session bus and the Wayland socket of that user.
, sessionUser ? null
}:
{ pkgs, lib, config, ... }:
# `mycli hostd serve`: small HTTP API (status / suspend / power off / file
# transfer / desktop notifications / push token registry) used by the myremote
# phone app. Devices register their FCM token with POST /push-token and `mycli
# mobile send` on this host pushes text to them. Requests are HMAC-signed with a
# token that the service generates on first start, print it with `sudo mycli hostd state`.
# See api/hostd/auth.go for the signing scheme.
let
  mycli = pkgs.callPackage ../packages/mycli.nix { };
  notify = sessionUser != null;
  user = if notify then sessionUser else "hostd";
in
{
  users.groups.hostd = { };
  users.users.hostd = lib.mkIf (!notify) {
    isSystemUser = true;
    group = "hostd";
  };

  # Needed for `sudo mycli hostd state`, root does not have the home-manager profile.
  environment.systemPackages = [ mycli ];

  systemd.services.hostd = {
    description = "Remote host management API";
    after = [ "network.target" ];
    wantedBy = [ "multi-user.target" ];
    # systemctl + busctl, suspend and power off go through logind. systemd-run
    # starts the browser outside of the sandbox with this PATH: xdg-open accepts
    # the default browser only when its Exec (`firefox`, not absolute) is on it,
    # otherwise it silently falls back to any browser it finds.
    path = [ pkgs.systemd ] ++ lib.optionals notify [
      pkgs.libnotify
      pkgs.wl-clipboard
      pkgs.xdg-utils
      "/etc/profiles/per-user/${user}"
      "/run/current-system/sw"
    ];
    serviceConfig = {
      ExecStart = lib.concatStringsSep " " (
        [ "${mycli}/bin/mycli" "hostd" "serve" "--listen" ":${toString port}" ]
        ++ lib.optionals (filesDir != null) [ "--files-dir" filesDir ]
        ++ lib.optional notify "--notify"
      );
      Restart = "on-failure";
      RestartSec = 5;
      # With sessionUser the session sockets under /run/user are reachable through
      # the read-only /run; the Go side fills in the environment pointing at them.
      User = user;
      Group = "hostd";
      # token
      StateDirectory = "hostd";
      StateDirectoryMode = "0700";

      NoNewPrivileges = true;
      ProtectSystem = "strict";
      ReadWritePaths = lib.optional (filesDir != null) filesDir;
      # true would also hide /run/user, where the session bus and Wayland sockets live
      ProtectHome = if notify then "read-only" else true;
      PrivateTmp = true;
      PrivateDevices = true;
      ProtectKernelTunables = true;
      ProtectKernelModules = true;
      ProtectControlGroups = true;
      RestrictNamespaces = true;
      LockPersonality = true;
      CapabilityBoundingSet = "";
      # AF_UNIX for D-Bus
      RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" ];
    };
  };

  # Transferred files have to be usable by both sides: the directory is setgid,
  # so files dropped there by the user end up in the hostd group, uploads are
  # created group-writable and the user is in that group. The token is not
  # exposed by that, /var/lib/hostd is 0700.
  systemd.tmpfiles.rules = lib.optionals (filesDir != null) (
    [ "d ${filesDir} 2770 hostd hostd -" ]
    ++ lib.optional (filesLink != null) "L ${filesLink} - - - - ${filesDir}"
  );
  users.users.${config.myconfig.username}.extraGroups = lib.optional (filesDir != null) "hostd";

  # hostd is not root and has no session, so logind asks polkit. The
  # multiple-sessions variants are needed whenever someone is logged in
  # (desktop session, ssh). Ignoring inhibitors is deliberately not allowed.
  security.polkit.enable = true;
  security.polkit.extraConfig = ''
    polkit.addRule(function(action, subject) {
      var allowed = [
        "org.freedesktop.login1.suspend",
        "org.freedesktop.login1.suspend-multiple-sessions",
        "org.freedesktop.login1.power-off",
        "org.freedesktop.login1.power-off-multiple-sessions"
      ];
      if (subject.user == "${user}" && allowed.indexOf(action.id) >= 0) {
        return polkit.Result.YES;
      }
    });
  '';

  networking.firewall.interfaces.${lanInterface}.allowedTCPPorts = [ port ];
  networking.firewall.interfaces."tailscale0".allowedTCPPorts = [ port ];
}
