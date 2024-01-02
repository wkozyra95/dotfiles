inputs:
{ pkgs, ... }:
{
  nix.settings.experimental-features = [ "nix-command" "flakes" ];
  programs.direnv.enable = true;
  programs.zsh = {
    enable = true;
    ohMyZsh = {
      enable = true;
      plugins = [ "git" "common-aliases" "cp" "docker" "golang" "vi-mode" "vim-interaction" ];
      theme = "bira";
    };
  };
  environment.systemPackages = with pkgs; [
    wget
    sway
    git
    git-crypt
    wget
    curl
    gnumake
    unzip
    jq
    killall
    cmake
    ripgrep
    python3Packages.pygments # needed by oh-my-zsh plugin
  ];
}
