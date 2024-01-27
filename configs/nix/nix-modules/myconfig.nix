myconfig:
{ lib, config, ... }:
let
  sharedModule = (import ../hm-modules/myconfig.nix myconfig);
in
{
  imports = [ sharedModule ];

  options.myconfig = {
    hm-modules = lib.mkOption {
      type = lib.types.listOf lib.types.anything;
      default = [ ];
    };
  };

  config = {
    myconfig.hm-modules = [
      sharedModule
    ];
  };
}
