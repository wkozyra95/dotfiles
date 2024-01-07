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
  };
}
