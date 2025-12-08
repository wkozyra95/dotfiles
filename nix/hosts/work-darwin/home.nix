systemModules:
{ pkgs, config, ... }:
{
  imports = [
    ../../hm-modules/common.nix
    ../../hm-modules/git.nix
    ../../hm-modules/vim.nix
    ../../hm-modules/neovim.nix
    ../../hm-modules/dotfiles.nix
  ] ++ systemModules;

  config = {
    home.packages = with pkgs; [
      ueberzugpp
    ];

    programs.ranger.extraConfig = ''
      set preview_images true
      set preview_images_method ueberzug
    '';

    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
