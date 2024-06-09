{ config, ... }:
{
  virtualisation.docker.enable = true;
  users.users.${config.myconfig.username} = {
    extraGroups = [ "docker" ];
  };
}
