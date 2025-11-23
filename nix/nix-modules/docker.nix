{ config, pkgs, ... }:
{
  virtualisation.docker.enable = true;
  environment.systemPackages = [ pkgs.docker-compose ];
  users.users.${config.myconfig.username} = {
    extraGroups = [ "docker" ];
  };
}
