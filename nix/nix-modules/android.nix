{ pkgs, config, ... }:
{
  programs.adb.enable = true;
  services.udev.packages = [
    pkgs.android-udev-rules
  ];
  users.users.${config.myconfig.username} = {
    extraGroups = [ "adbusers" ];
  };
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
