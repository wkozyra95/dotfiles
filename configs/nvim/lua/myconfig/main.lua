vim.g.mapleader = " "

require("myconfig.globals")
require("myconfig.base").apply()
require("myconfig.statusline")
require("myconfig.options")
require("myconfig.format").apply()
require("myconfig.spell").apply()
require("myconfig.filetype").apply()
require("myconfig.snippets").apply()
require("myconfig.telescope").apply()
require("myconfig.git").apply()

require("myconfig.workspaces").apply(
    function()
        require("myconfig.actions")
        require("myconfig.keymap")
        require("myconfig.treesitter").apply()
        require("myconfig.lsp").apply()
        require("myconfig.playground.node").apply()
        require("myconfig.playground.docker").apply()
        require("myconfig.noice").apply()
        require("myconfig.db").apply()
        require("myconfig.rest").apply()
        require("myconfig.amp").apply()
    end
)
