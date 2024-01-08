{ pkgs, config, ... }:
let
  mkOutOfStoreSymlink = config.lib.file.mkOutOfStoreSymlink;
  dotfilesSymlink = path:
    config.lib.file.mkOutOfStoreSymlink "${config.home.homeDirectory}/.dotfiles/${path}";
in
{
  programs.direnv.enable = true;
  programs.zsh = {
    enable = true;
    oh-my-zsh = {
      enable = true;
    };
  };

  home.file = {
    ".zshrc".source = dotfilesSymlink "configs/zshrc";
    ".vimrc".source = dotfilesSymlink "configs/vimrc";
    ".ideavimrc".source = dotfilesSymlink "configs/ideavimrc";
    ".docker".source = dotfilesSymlink "configs/docker";

    ".config/sway".source = dotfilesSymlink "configs/sway";
    ".config/i3".source = dotfilesSymlink "configs/i3";
    ".config/alacritty.yml".source = dotfilesSymlink "configs/alacritty.yml";
    ".config/nvim".source = dotfilesSymlink "configs/nvim";
    ".config/direnv".source = dotfilesSymlink "configs/direnv";

    "notes".source = mkOutOfStoreSymlink
      "${config.home.homeDirectory}/.dotfiles-private/notes";
    ".dotfiles/configs/nvim/spell".source = mkOutOfStoreSymlink
      "${config.home.homeDirectory}/.dotfiles-private/nvim/spell";
  };

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
    nixpkgs-fmt
    tree-sitter
    (pkgs.callPackage ../packages/lua-code-format.nix { })
  ];
}
