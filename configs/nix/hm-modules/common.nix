{ pkgs, config, ... }:
let
  mkOutOfStoreSymlink = config.lib.file.mkOutOfStoreSymlink;
  dotfilesSymlink = path:
    config.lib.file.mkOutOfStoreSymlink "${config.home.homeDirectory}/.dotfiles/${path}";
in
{
  programs.direnv = {
    enable = true;
    enableZshIntegration = true;
    nix-direnv.enable = true;
    config = {
      whitelist.prefix = [
        "${config.home.homeDirectory}/.dotfiles"
      ];
    };
  };
  programs.zsh = {
    enable = true;
    history = {
      save = 1000000;
      size = 1000000;
      share = true;
    };
    shellAliases = {
      g = "git";
      git = "git";
      ggpush = "git push --set-upstream origin $(git_current_branch)";
    };
    initExtra = ''
      function try_source() {
          test -s $1 && source $1
      }
      try_source $HOME/.zshrc.secrets
      try_source $HOME/.cache/mycli/completion/zsh_setup
    '';
    oh-my-zsh = {
      enable = true;
      plugins = [
        "git"
        "common-aliases"
        "docker"
        "golang"
        "vi-mode"
      ];
      custom = "$HOME/.dotfiles/configs/zsh";
      theme = "bira";
    };
  };
  programs.fzf.enable = true;

  home.file = {
    ".ideavimrc".source = dotfilesSymlink "configs/ideavimrc";
    ".docker".source = dotfilesSymlink "configs/docker";

    ".config/sway".source = dotfilesSymlink "configs/sway";
    ".config/i3".source = dotfilesSymlink "configs/i3";
    ".config/alacritty.yml".source = dotfilesSymlink "configs/alacritty.yml";

    "notes".source = mkOutOfStoreSymlink
      "${config.home.homeDirectory}/.dotfiles-private/notes";
    ".dotfiles/configs/nvim/spell".source = mkOutOfStoreSymlink
      "${config.home.homeDirectory}/.dotfiles-private/nvim/spell";
  };

  programs.neovim.enable = true;

  home.sessionVariables = {
    EDITOR = "nvim";
    CURRENT_ENV = config.myconfig.env;
  };

  home.packages = with pkgs; [
    git-crypt
    wget
    curl
    unzip
    jq
    btop
    rsync
    killall
    ripgrep
    python3Packages.pygments # needed by oh-my-zsh plugin
    tree-sitter
    silver-searcher

    # LSP
    nodejs_18
    nil
    sumneko-lua-language-server
    nodePackages.typescript-language-server
    vscode-langservers-extracted
    efm-langserver
    elixir_ls

    # Custom
    (pkgs.callPackage ../packages/lua-code-format.nix { })
    (pkgs.callPackage ../packages/mycli.nix { })
  ];

  programs.gpg.enable = true;
  services.gpg-agent = {
    enable = true;
    pinentryFlavor = "curses";
    enableSshSupport = true;
    enableExtraSocket = true;
  };
}
