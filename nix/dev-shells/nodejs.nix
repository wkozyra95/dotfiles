{ pkgs, ... }:

pkgs.mkShell {
  packages = with pkgs; [
    nodejs
    corepack
    typescript
    typescript-language-server
    prettier
    eslint
  ];
}
