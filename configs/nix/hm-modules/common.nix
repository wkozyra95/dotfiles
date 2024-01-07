{ pkgs, ... }:
{
  programs.direnv.enable = true;
  #programs.zsh = {
  #  enable = true;
  #  oh-my-zsh = {
  #    enable = true;
  #  };
  #};

  programs.neovim.enable = true;

  home.sessionVariables = {
    EDITOR = "nvim";
  };

  home.packages = with pkgs; [
    wget
    git
    git-crypt
    git-lfs
    diff-so-fancy
    wget
    curl
    unzip
    jq
    killall
    ripgrep
    python3Packages.pygments # needed by oh-my-zsh plugin
    vim
    zsh
    (pkgs.callPackage ../packages/lua-code-format.nix {})
  ];

  nixpkgs.config.allowUnfree = true;
  nix = {
    settings = {
      experimental-features = [ "nix-command" "flakes" ];
      allowed-users = [ "@wheel" "@sudo" ];
    };
  };
}
