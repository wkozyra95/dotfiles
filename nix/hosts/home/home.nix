systemModules:
{ pkgs, config, ... }:
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
    home.packages = with pkgs; [
      bitwarden-cli
      gh
      obs-studio
      prusa-slicer
    ];

    myconfig = {
      git.signingKey = "35DF 8DFA D0E7 1E39 F047 BD01 AE51 A568 2B78 648C";
    };

    programs.gpg.enable = true;
    services.gpg-agent = {
      enable = true;
      pinentry = {
        package = pkgs.pinentry-curses;
      };
      enableSshSupport = true;
      enableExtraSocket = true;
    };

    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
