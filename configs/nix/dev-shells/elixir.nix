{ pkgs, ... }:

pkgs.mkShell {
  packages = with pkgs; [
    elixir
  ];
}
