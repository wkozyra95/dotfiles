systemModules:
{ pkgs, config, ... }:
{
  imports = [
    ../common.nix
    ../hm-modules/common.nix
    ../hm-modules/common-desktop.nix
    ../hm-modules/git.nix
    ../hm-modules/vim.nix
    ../hm-modules/neovim.nix
    ../hm-modules/dotfiles.nix
  ] ++ systemModules;

  config = {
    home.packages = with pkgs; [
      bitwarden-cli
      gh
    ];

    myconfig = {
      git.signingKey = "35DF 8DFA D0E7 1E39 F047 BD01 AE51 A568 2B78 648C";
    };
    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
