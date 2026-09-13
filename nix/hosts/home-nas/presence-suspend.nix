{ pkgs, lib, ... }:
# Suspend the NAS when the workstation has been off the LAN for a while.
#
# Wake it with a magic packet to 00:d8:61:7a:45:c4 (e.g. `wakeonlan` from the
# workstation or a phone app); every wake gets at least graceMin minutes
# before the next check can put it back to sleep.
#
# WoL needs "Resume By PCI-E Device" enabled in the BIOS and
# networking.interfaces.enp39s0.wakeOnLan (system.nix).
#
# Escape hatch: `touch /run/keep-awake` blocks suspending until reboot.
let
  workstationIp = "192.168.100.9"; # give it a DHCP reservation in the router
  graceMin = 30;                   # minutes unreachable (or since wake) before suspending
  stateDir = "/run/presence-suspend";
in
{
  systemd.services.presence-suspend = {
    description = "Suspend when the workstation has been unreachable for ${toString graceMin} min";
    path = [ pkgs.iputils pkgs.coreutils pkgs.systemd ];
    serviceConfig.Type = "oneshot";
    script = ''
      set -eu
      mkdir -p ${stateDir}
      now=$(date +%s)
      last=${stateDir}/last-seen
      [ -f "$last" ] || echo "$now" > "$last"

      if [ -e /run/keep-awake ]; then
        echo "keep-awake present, not suspending"; exit 0
      fi
      if ping -c 2 -W 2 ${workstationIp} >/dev/null 2>&1; then
        echo "$now" > "$last"; exit 0
      fi
      since=$(( now - $(cat "$last") ))
      if [ "$since" -ge $(( ${toString graceMin} * 60 )) ]; then
        echo "workstation unreachable for $((since/60)) min, suspending"
        systemctl suspend
      else
        echo "workstation unreachable for $((since/60)) min, waiting"
      fi
    '';
  };

  systemd.timers.presence-suspend = {
    wantedBy = [ "timers.target" ];
    timerConfig = {
      OnBootSec = "5min";
      OnUnitActiveSec = "5min";
    };
  };

  # Reset the clock on boot and on every resume so a fresh wake (WoL from a
  # phone, or the power button) always gets the full grace period.
  systemd.services.presence-suspend-reset = {
    description = "Reset presence-suspend grace period after boot/resume";
    wantedBy = [ "multi-user.target" "post-resume.target" ];
    after = [ "post-resume.target" ];
    serviceConfig.Type = "oneshot";
    script = ''
      mkdir -p ${stateDir}
      date +%s > ${stateDir}/last-seen
    '';
  };
}
