overlay:
{ pkgs, ... }:
{
  environment.variables.EDITOR = "nvim";
  programs.neovim = {
    enable = true;
  };
  nixpkgs = {
    overlays = [overlay];
  };
}
