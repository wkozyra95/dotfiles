# Common options, This module can be used as nixos/nix-darwin/home-manger module.
{
  nixpkgs.config.allowUnfree = true;
  nix = {
    settings = {
      experimental-features = [ "nix-command" "flakes" ];
    };
  };
}
