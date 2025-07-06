systemModules:
{ pkgs, config, lib, ... }:
{
  imports = [
    ../../hm-modules/common.nix
    ../../hm-modules/common-desktop.nix
    ../../hm-modules/git.nix
    ../../hm-modules/vim.nix
    ../../hm-modules/neovim.nix
    ../../hm-modules/dotfiles.nix
  ] ++ systemModules;

  config = {
    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
