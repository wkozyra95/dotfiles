{ nixpkgs, ... }:
nixpkgs.lib.nixosSystem {
  system = "x86_64-linux";
  modules = [
    "${nixpkgs}/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix"
    "${nixpkgs}/nixos/modules/installer/cd-dvd/channel.nix"
    ({ pkgs, ... }: {
      isoImage.contents = [
        {
          source = ../..;
          target = "/dotfiles";
        }
      ];
      programs.zsh.enable = true;
      programs.zsh.ohMyZsh.enable = true;
      environment.systemPackages = with pkgs; [
        neovim
        curl
        wget
        git
        (pkgs.callPackage ../packages/mycli.nix { })
      ];
    })
  ];
}
