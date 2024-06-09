systemModules:
{ pkgs, config, ... }:
{
  imports = [
    ../../common.nix
    ../../hm-modules/common.nix
    ../../hm-modules/common-desktop.nix
    ../../hm-modules/git.nix
    ../../hm-modules/vim.nix
    ../../hm-modules/neovim.nix
    ../../hm-modules/dotfiles.nix
  ] ++ systemModules;

  config = {
    programs.gpg.enable = true;
    services.gpg-agent = {
      enable = true;
      pinentryFlavor = "curses";
      enableSshSupport = true;
      enableExtraSocket = true;
    };

    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
