vim.opt.rtp:prepend(vim.fn.stdpath("data") .. "/lazy/lazy.nvim")
local snapshot_path = "~/.dotfiles/configs/nvim/lazy-lock.json"

vim.api.nvim_create_user_command(
    "PackagesInstall", function() vim.api.nvim_command("Lazy restore") end, {nargs = 0}
)
vim.api.nvim_create_user_command(
    "PackagesUpdate", function() vim.api.nvim_command("Lazy update") end, {nargs = 0}
)

require("lazy").setup(
    {
        "nvim-lua/popup.nvim",
        "kyazdani42/nvim-web-devicons",

        "nvim-lua/plenary.nvim",
        "nvim-telescope/telescope.nvim",
        {
            "nvim-telescope/telescope-fzf-native.nvim",
            build =
            "cmake -S. -Bbuild -DCMAKE_BUILD_TYPE=Release && cmake --build build --config Release && cmake --install build --prefix build",
        },
        {"nvim-telescope/telescope-file-browser.nvim"},
        {
            "folke/noice.nvim",
            dependencies = {"MunifTanjim/nui.nvim",
                "rcarriga/nvim-notify"}
        },
        {"TimUntersberger/neogit",                    dependencies = "nvim-lua/plenary.nvim"},
        {"sindrets/diffview.nvim",                    dependencies = "nvim-lua/plenary.nvim"},
        "airblade/vim-gitgutter",
        "tpope/vim-fugitive",

        "tpope/vim-dadbod",
        "kristijanhusak/vim-dadbod-ui",

        "neovim/nvim-lspconfig",
        "hrsh7th/nvim-cmp",
        "hrsh7th/cmp-nvim-lsp",
        "hrsh7th/cmp-buffer",
        "hrsh7th/cmp-path",

        "L3MON4D3/LuaSnip",
        "saadparwaiz1/cmp_luasnip",

        "onsails/lspkind-nvim",

        "nanotee/luv-vimdocs",  -- nvim event loop
        "milisims/nvim-luaref", -- lua bultin

        "tpope/vim-endwise",
        "numToStr/Comment.nvim",

        {
            "ellisonleao/gruvbox.nvim",
            priority = 1000,
            lazy = false,
            config = function() vim.cmd.colorscheme("gruvbox") end,
        },
        "gorodinskiy/vim-coloresque", -- color preview in css

        {"nvim-treesitter/nvim-treesitter"},
        "nvim-treesitter/playground",
        "nvim-treesitter/nvim-treesitter-context",

    }, {lockfile = snapshot_path, install = {missing = true, colorscheme = {"gruvbox"}}}
)
