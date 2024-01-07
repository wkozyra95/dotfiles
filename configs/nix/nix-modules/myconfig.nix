{ username }:
{ config, lib, pkgs, ... }: {
  options.myconfig = {
    username = lib.mkOption {
      type = lib.types.str;
      default = username;
    };
  };
}
