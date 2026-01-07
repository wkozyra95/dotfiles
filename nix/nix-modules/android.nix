{ config, ... }:
{
  programs.adb.enable = true;
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
