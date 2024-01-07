{ nixpkgs, home-manager, myConfigModule, overlays }:

nixpkgs.lib.nixosSystem {
  system = "x86_64-linux";

  modules = [
    home-manager.nixosModules.home-manager
    myConfigModule
    (import ./filesystems.nix)
    (import ./system.nix)
    (import ./boot.nix)
    (import ../nix-modules/sway.nix)
    (import ../nix-modules/docker.nix)
    (import ../nix-modules/steam.nix)
    ({ config, lib, pkgs, ... }: {
      nixpkgs.overlays = overlays;
      home-manager = {
        useGlobalPkgs = true;
        useUserPackages = true;
        users.${config.myconfig.username} = (import ./home.nix);
      };
    })
  ];
}
