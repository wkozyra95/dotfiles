set exrc
set secure

filetype plugin indent on

let mapleader = "<space>"

noremap <C-e> :NERDTreeToggle<CR>
noremap <space><space> :NERDTreeFind<CR>
noremap <bs> <C-^>
noremap <C-p> :FZF<CR>
noremap <S-h> :wa<CR>gT
noremap <S-l> :wa<CR>gt
noremap <C-f> :Ag<Space>
noremap <Leader>s :w<CR>
noremap <Leader>= gg=G
noremap <C-n> :tab split<CR>

set number relativenumber
set hlsearch
set incsearch
set laststatus=2
set mouse=a
set backspace=indent,eol,start
let $FZF_DEFAULT_COMMAND = 'ag --ignore-case --ignore .git -l -g ""'
set nofoldenable
set termguicolors


syntax enable
set background=dark
try
    colorscheme gruvbox
catch
endtry

set tabstop=4 shiftwidth=4 softtabstop=4 expandtab smarttab
augroup indent
    autocmd!
    autocmd FileType python setlocal ts=4 sw=4 sts=4 et smarttab
    autocmd FileType lua setlocal ts=4 sw=4 sts=4 et smarttab
    autocmd FileType sh setlocal ts=2 sw=2 sts=2 et smarttab
    autocmd FileType javascript setlocal ts=2 sw=2 sts=2 et smarttab
    autocmd FileType json setlocal ts=2 sw=2 sts=2 et smarttab
    autocmd FileType typescript setlocal ts=2 sw=2 sts=2 et smarttab
    autocmd FileType typescriptreact setlocal ts=2 sw=2 sts=2 et smarttab
    autocmd FileType css setlocal ts=2 sw=2 sts=2 et smarttab
    autocmd FileType scss setlocal ts=2 sw=2 sts=2 et smarttab
    autocmd FileType yaml setlocal ts=2 sts=2 sw=2 expandtab
    autocmd FileType plist setlocal ts=4 sts=4 sw=4 expandtab
augroup END

let g:NERDTreeShowHidden = 1
let NERDTreeIgnore=['\.py[cd]$', '\~$', '\.swo$', '\.swp$', '^\.git$', '^\.hg$', '^\.svn$', '\.bzr$', 'node_modules']
let NERDTreeMouseMode=2

"   VIM_AIRLINE
let g:airline_powerline_fonts=1
let g:airline_theme = 'gruvbox'
