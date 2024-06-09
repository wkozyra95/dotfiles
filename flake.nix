{
  description = "Setup entire environment.";

  inputs = {
    # NixOS
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";

    # Nix darwin
    nixpkgs-darwin.url = "github:NixOS/nixpkgs/nixpkgs-24.05-darwin";
    nix-darwin = {
      url = "github:LnL7/nix-darwin";
      inputs.nixpkgs.follows = "nixpkgs-darwin";
    };

    home-manager = {
      url = "github:nix-community/home-manager/release-24.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    neovim-nightly-overlay = {
      url = "github:nix-community/neovim-nightly-overlay/7b5ca2486bba58cac80b9229209239740b67cf90";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, flake-parts, nixpkgs, nixpkgs-unstable, nix-darwin, home-manager, ... }@inputs:
    let
      perSystemConfig = flake-parts.lib.mkFlake { inherit inputs; } {
        systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];
        perSystem = { config, self', inputs', pkgs, system, lib, ... }@args:
          {
            _module.args.pkgs = import nixpkgs-unstable {
              inherit system;
              config = { allowUnfree = true; };
            };
            devShells = import ./nix/dev-shells args;
            formatter = pkgs.nixpkgs-fmt;
          };
      };
      overlays = [ inputs.neovim-nightly-overlay.overlays.default ];
      opts = {
        inherit nixpkgs home-manager overlays nixpkgs-unstable;
      };
    in
    {
      nixosConfigurations = {
        # sudo nixos-rebuild switch --flake ".#home"
        home = (import ./nix/hosts/home opts);
        usbdrive = (import ./nix/hosts/usbdrive opts);
        # Build installer ISO
        # nix build .#nixosConfigurations.iso-installer.config.system.build.isoImage
        iso-installer = (import ./nix/hosts/iso.nix opts);
        # Build vm
        # nix build .#nixosConfigurations.dev-vm.config.system.build.vm
        # Run vm
        # ./result/bin/run-dev-vm
        dev-vm = (import ./nix/hosts/nixos-vm opts);
      };
      darwinConfigurations = {
        # First install:
        # nix run nix-darwin -- switch --flake ".#work-mac" 
        # Rebuild:
        # darwin-rebuild switch --flake ".#work-mac"
        work-mac = (import ./nix/hosts/work-darwin {
          inherit nix-darwin home-manager overlays nixpkgs-unstable;
        });
      };
      homeConfigurations = {
        # Work desktop config
        # First install:
        # nix run home-manager/release-23.11 -- switch --flake ".#work"
        # Rebuild:
        # home-manger switch --flake ".#work"
        work = (import ./nix/hosts/work-arch opts);
        # Config for VM for development non-nixos
        dev-vm = (import ./nix/hosts/dev-vm opts);
      };
      devShells = perSystemConfig.devShells;
      packages = perSystemConfig.packages;
      formatter = perSystemConfig.formatter;
    };
}
