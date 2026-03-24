{ pkgs, ... }:

pkgs.mkShell {
  packages = with pkgs; [
    nodejs
    corepack
    typescript
    nodePackages.typescript-language-server
    nodePackages.prettier
    nodePackages.eslint
  ];
}
