{ pkgs, config, ... }:
{
  programs.neovim = {
    enable = true;
    defaultEditor = true;
    withPython3 = false;
    withRuby = false;
    initLua = ''
      vim.opt.rtp:prepend("${config.home.homeDirectory}/.dotfiles/configs/nvim")
      require("myconfig.main")
    '';
    plugins = with pkgs.vimPlugins; [
      popup-nvim
      nvim-web-devicons
      plenary-nvim

      telescope-nvim
      telescope-fzy-native-nvim
      telescope-file-browser-nvim

      noice-nvim
      nui-nvim
      nvim-notify

      neogit
      diffview-nvim
      vim-gitgutter
      vim-fugitive

      vim-dadbod
      vim-dadbod-ui
      vim-dadbod-completion

      rest-nvim

      nvim-lspconfig

      nvim-cmp
      cmp-nvim-lsp
      cmp-buffer
      cmp-path

      luasnip
      cmp_luasnip

      lspkind-nvim

      vim-endwise
      comment-nvim

      gruvbox-nvim

      nvim-treesitter.withAllGrammars
      nvim-treesitter-context

      amp-nvim
      claudecode-nvim
      snacks-nvim
    ];
    # inotifywait is used by nvim as a backend for LSP file watching
    extraPackages = pkgs.lib.optionals pkgs.stdenv.isLinux [ pkgs.inotify-tools ];
    extraLuaPackages = pkgs: [
      pkgs.lua-curl
      pkgs.xml2lua
      pkgs.mimetypes
      pkgs.nvim-nio
    ];
  };
}
