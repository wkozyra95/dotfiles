{
  description = "Setup entire environment.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    neovim-nightly-overlay.url = "github:nix-community/neovim-nightly-overlay";
  };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      devShell = system: {
        name = system;
        value = import ./develop.nix {
          pkgs = nixpkgs.legacyPackages.${system};
          inherit system;
        };
      };
    in
    {
      nixosConfigurations = {
        # sudo nixos-rebuild switch --flake ".#home"
        home = (import ./configs/nix/home inputs);
      };

      # nix develop
      devShells = builtins.listToAttrs (builtins.map(devShell) [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ]);
    };
}
