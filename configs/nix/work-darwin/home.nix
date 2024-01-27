systemModules:
{ pkgs, config, ... }:
{
  imports = [
    ../common.nix
    ../hm-modules/common.nix
    ../hm-modules/git.nix
  ] ++ systemModules;

  config = {
    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
