{
  description = "Setup entire environment.";

  inputs = {
    # NixOS
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";

    # Nix darwin
    nixpkgs-darwin.url = "github:NixOS/nixpkgs/nixpkgs-23.11-darwin";
    nix-darwin = {
      url = "github:LnL7/nix-darwin";
      inputs.nixpkgs.follows = "nixpkgs-darwin";
    };

    home-manager = {
      url = "github:nix-community/home-manager/release-23.11";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    neovim-nightly-overlay.url = "github:nix-community/neovim-nightly-overlay";
  };

  outputs = { self, flake-parts, nixpkgs, nix-darwin, home-manager, ... }@inputs:
    let
      perSystemConfig = flake-parts.lib.mkFlake { inherit inputs; } {
        systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];
        perSystem = { config, self', inputs', pkgs, system, lib, ... }@args:
          {
            devShells = import ./configs/nix/dev-shells args;
            formatter = pkgs.nixpkgs-fmt;
          };
      };
      overlays = [ inputs.neovim-nightly-overlay.overlay ];
    in
    {
      nixosConfigurations = {
        # sudo nixos-rebuild switch --flake ".#home"
        home = (import ./configs/nix/home {
          inherit nixpkgs home-manager overlays;
        });
        # Build installer ISO
        # nix build .#nixosConfigurations.iso-installer.config.system.build.isoImage
        iso-installer = (import ./configs/nix/iso.nix {
          inherit nixpkgs;
        });
        # Build vm
        # nix build .#nixosConfigurations.dev.config.system.build.vm
        # Run vm
        # ./result/bin/run-dev-vm
        dev-vm = (import ./configs/nix/dev-vm {
          inherit nixpkgs home-manager overlays;
        });
      };
      darwinConfigurations = {
        # First install:
        # nix run nix-darwin -- switch --flake ".#work-mac" 
        # Rebuild:
        # darwin-rebuild switch --flake ".#work-mac"
        work-mac = (import ./configs/nix/work-darwin {
          inherit nix-darwin home-manager overlays;
        });
      };
      homeConfigurations = {
        # Work desktop config
        # First install:
        # nix run home-manager/release-23.11 -- switch --flake ".#work"
        # Rebuild:
        # home-manger switch --flake ".#work"
        work = (import ./configs/nix/work-arch {
          inherit home-manager overlays;
          pkgs = nixpkgs.legacyPackages.x86_64-linux;
        });
      };
      devShells = perSystemConfig.devShells;
      packages = perSystemConfig.packages;
      formatter = perSystemConfig.formatter;
    };
}
