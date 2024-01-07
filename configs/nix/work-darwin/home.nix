{ pkgs, ...}:
{
  imports = [
    ../hm-modules/common.nix
    ../hm-modules/languages
  ];

  config = {
    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
