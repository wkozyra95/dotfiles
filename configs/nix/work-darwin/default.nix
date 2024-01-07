{ nix-darwin, home-manager, overlays }:

nix-darwin.lib.darwinSystem {
  system = "x86_64-darwin";

  modules = [
    home-manager.darwinModules.home-manager
    (import ../nix-modules/myconfig.nix {
      username = "wojciechkozyra";
      email = "wojciechkozyra@swmansion.com";
    })
    ({ config, lib, pkgs, ... }: {
      home-manager = {
        useGlobalPkgs = true;
        useUserPackages = true;
        users.${config.myconfig.username} = (import ./home.nix);
      };

      time.timeZone = "Europe/Warsaw";

      networking.hostName = "wojtek-mac";

      nixpkgs = {
        hostPlatform = lib.mkDefault "x86_64-darwin";
        overlays = overlays;
      };

      services.nix-daemon.enable = true;
      users.users.${config.myconfig.username} = {
        home = "/Users/${config.mycconfig.username}";
      };

      system.stateVersion = 4;
    })
  ];
}
