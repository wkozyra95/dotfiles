{
  description = "Setup entire environment.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    neovim-nightly-overlay.url = "github:nix-community/neovim-nightly-overlay";
  };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      devShell = system: {
        name = system;
        value = import ./develop.nix {
          pkgs = nixpkgs.legacyPackages.${system};
          inherit system;
        };
      };
      packages = system: {
        name = system;
        value = (
          let
            callPackage = nixpkgs.legacyPackages.${system}.callPackage;
          in
          {
            lua-code-format = callPackage ./configs/nix/packages/lua-code-format.nix {};
          }
        );
      };
    in
    {
      nixosConfigurations = {
        # sudo nixos-rebuild switch --flake ".#home"
        home = (import ./configs/nix/home {
          inputs = inputs;
          customPackages = builtins.attrValues inputs.self.packages.x86_64-linux;
        });
      };

      # nix develop
      devShells = builtins.listToAttrs (builtins.map(devShell) systems);

      packages = builtins.listToAttrs (builtins.map(packages) systems);
    };
}
