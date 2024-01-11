{ username, email }:
{ config, lib, pkgs, ... }: {
  options.myconfig = {
    username = lib.mkOption {
      type = lib.types.str;
      default = username;
    };
    email = lib.mkOption {
      type = lib.types.str;
      default = email;
    };
    hm-modules = lib.mkOption {
      type = lib.types.listOf lib.types.anything;
      default = [ ];
    };
  };
}
