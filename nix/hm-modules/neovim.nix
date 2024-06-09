{ unstable, config, ... }:
{
  programs.neovim = {
    enable = true;
    defaultEditor = true;
    extraLuaConfig = ''
      vim.opt.rtp:prepend("${config.home.homeDirectory}/.dotfiles/configs/nvim")
      require("myconfig.main")
    '';
    plugins = with unstable.vimPlugins; [
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
      playground
      nvim-treesitter-context
    ];
    extraLuaPackages = pkgs: [
      pkgs.lua-curl
      pkgs.xml2lua
      pkgs.mimetypes
      pkgs.nvim-nio
    ];
  };
}
