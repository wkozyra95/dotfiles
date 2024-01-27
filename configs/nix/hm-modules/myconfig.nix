{ username, email }:
{ lib, ... }: {
  # This modules is used both as NixOS, nix-darwin, and Home Manager module, so
  # it should only rely on API available for both.
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
