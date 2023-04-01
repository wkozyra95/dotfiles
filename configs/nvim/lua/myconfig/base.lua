local module = {}

local icons = require("nvim-web-devicons")

function module.apply()
    vim.cmd.filetype("plugin indent on")
    vim.cmd.syntax("enable")
    vim.g.AutoPairs = {["("] = ")",["["] = "]",["{"] = "}",["`"] = "`",["```"] = "```"}

    vim.g.surround_no_mappings = 1
    local group = vim.api.nvim_create_augroup("main", {clear = true})
    vim.api.nvim_create_autocmd(
        "TextYankPost", {
            group = group,
            pattern = "*",
            callback = function()
                vim.highlight.on_yank({higroup = "IncSearch", timeout = 1000})
            end,
        }
    )
    icons.setup {
        override = {xml = {icon = "", color = "#e37933", cterm_color = "173", name = "Xml"}},
        default = true,
    }
end

return module
