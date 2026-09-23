{ pkgs, ... }:
{
  # programs.adb was removed in 26.05; systemd handles the udev uaccess rules,
  # only the adb binary is needed.
  environment.systemPackages = [ pkgs.android-tools ];
  myconfig.hm-modules = [
    (
      { custom, ... }:
      {
        home.packages = with custom.unstable; [
          android-studio
        ];
      }
    )
  ];
}
