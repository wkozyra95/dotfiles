{ pkgs, ... }:
{
  programs.vim = {
    enable = true;
    plugins = with pkgs.vimPlugins; [
      nerdtree-git-plugin
      nerdtree
      vim-fugitive
      vim-gitgutter
      fzf-vim
      gruvbox
      vim-airline
      vim-airline-themes
    ];
    extraConfig = builtins.readFile ./vimrc.vim;
  };
}
