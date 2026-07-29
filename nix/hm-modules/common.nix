{ pkgs, config, lib, ... }:
let
  mkOutOfStoreSymlink = config.lib.file.mkOutOfStoreSymlink;
  dotfilesSymlink = path:
    config.lib.file.mkOutOfStoreSymlink "${config.home.homeDirectory}/.dotfiles/${path}";
in
{
  nix = {
    settings = {
      experimental-features = [ "nix-command" "flakes" ];
    };
  };
  programs.direnv = {
    enable = true;
    enableZshIntegration = true;
    nix-direnv.enable = true;
    config = {
      whitelist.prefix = [
        "${config.home.homeDirectory}/.dotfiles"
        "${config.home.homeDirectory}/smelter"
      ];
    };
  };
  programs.zsh = {
    enable = true;
    history = {
      save = 1000000;
      size = 1000000;
      # Append-only history. SHARE_HISTORY is deliberately OFF: it re-finds its
      # read position in the file by per-entry timestamp, and when many session
      # shells start at once — or a reseed collapses every entry onto a single
      # timestamp (which is what enabling EXTENDED_HISTORY on a timestamp-less
      # file did, see git history) — that positioning breaks. The result is both
      # runaway duplication (shells re-import and re-append what they already
      # read) and whole-file truncation (a shell with a partial in-memory view
      # rewrites the file). append=true makes the exit-save append rather than
      # rewrite, so a stale shell can never clobber the file; INC_APPEND_HISTORY
      # (set in initContent) persists each command immediately. extended is safe
      # now that share is off — it just records accurate per-command timestamps.
      share = false;
      append = true;
      extended = true;
    };
    shellAliases = {
      g = "git";
      ggpush = "git push --set-upstream origin $(git_current_branch)";
    };
    initContent = lib.mkMerge [
      ''
        function try_source() {
            test -s $1 && source $1
        }
        try_source $HOME/.zshrc.secrets
        try_source $HOME/.cache/mycli/completion/zsh_setup
      ''
      (lib.mkAfter ''
        # Persist each command to $HISTFILE the moment it runs, append-only. With
        # share=false + append=true above, no shell ever rewrites the whole file,
        # so a stale session can't truncate it. Not exposed as a home-manager
        # history option, so set it here; mkAfter keeps it after oh-my-zsh and the
        # generated history block (both of which would otherwise re-toggle opts).
        setopt INC_APPEND_HISTORY
        # ~/.zsh_history is chattr +a (kernel append-only) since the 2026-07-27
        # truncation: appends (O_WRONLY|O_APPEND) work, rewrites/renames get
        # EPERM. HIST_FCNTL_LOCK must stay off under +a — its lock open is
        # O_RDWR without O_APPEND, which EPERMs and makes zsh silently skip
        # every history write. The fallback $HISTFILE.LOCK locking is a
        # separate file, unaffected by the attribute.
        unsetopt HIST_FCNTL_LOCK
        # Even with append opts, zsh's exit-time save re-reads and REWRITES the
        # whole histfile — normally via the shared $HISTFILE.new + rename, which
        # is the race that truncated history on 2026-07-27 when two shells
        # exited at once. NO_HIST_SAVE_BY_COPY makes that rewrite in-place
        # instead, so chattr +a rejects it at open() — no .new litter, no
        # rename race; the only trace is one "failed to write history file"
        # stderr line per shell exit. All real data is already on disk from
        # INC_APPEND, so nothing is lost by the rewrite failing.
        unsetopt HIST_SAVE_BY_COPY
      '')
    ];
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
    ".config/waybar".source = dotfilesSymlink "configs/waybar";
    ".config/fuzzel".source = dotfilesSymlink "configs/fuzzel";
    ".config/i3".source = dotfilesSymlink "configs/i3";
    ".config/alacritty.yml".source = dotfilesSymlink "configs/alacritty.yml";
    ".config/alacritty.toml".source = dotfilesSymlink "configs/alacritty.toml";

    "notes".source = mkOutOfStoreSymlink
      "${config.home.homeDirectory}/.dotfiles-private/notes";
  };

  home.sessionPath = [
    "$HOME/.local/bin"
  ];

  programs.neovim.enable = true;

  home.sessionVariables = {
    EDITOR = "nvim";
    CURRENT_ENV = config.myconfig.env;
  };

  home.packages = with pkgs; [
    file
    lsof
    git-crypt
    docker-credential-helpers # provides docker-credential-secretservice; keeps registry tokens out of ~/.docker/config.json
    wget
    curl
    unzip
    jq
    btop
    htop
    rsync
    killall
    ripgrep
    python3Packages.pygments # needed by oh-my-zsh plugin
    tree-sitter
    silver-searcher
    ranger

    # LSP
    nodejs
    pnpm
    rustc
    cargo
    nil
    lua-language-server
    nodePackages.typescript-language-server
    vscode-langservers-extracted
    efm-langserver
    elixir-ls
    nixpkgs-fmt
    ltex-ls
    yaml-language-server
    clang-tools
    rust-analyzer
    # python
    basedpyright
    ty
    ruff

    # Custom
    (pkgs.callPackage ../packages/lua-code-format.nix { })
    (pkgs.callPackage ../packages/mycli.nix { })
  ];
}
