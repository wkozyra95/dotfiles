local module = {}

local format = require("myconfig.format")
local spell = require("myconfig.spell")
local snippets = require("myconfig.snippets")

local javascript = function()
    format.preset(2)
    spell.preset("ts")
    snippets.load("typescript", require("myconfig.lang.typescript_snippets"))
end

local handlers = {
    default = function() spell.preset("common") end,
    css = format.preset_2,
    scss = format.preset_2,
    sh = format.preset_2,
    cpp = format.preset_2,
    rust = function()
        format.preset(4)
        spell.preset("rust")
    end,
    go = function()
        format.preset(4)
        spell.preset("go")
        vim.opt_local.expandtab = false
        snippets.load("go", require("myconfig.lang.go_snippets"))
    end,
    elixir = require("myconfig.lang.elixir").apply,
    json = format.preset_2,
    lua = function()
        format.preset(4)
        spell.preset("lua")
        snippets.load("lua", require("myconfig.lang.lua_snippets"))
    end,
    markdown = function()
        format.preset(4)
        spell.strict_preset()
    end,
    gitcommit = spell.strict_preset,
    plist = format.preset_4,
    yaml = format.preset_2,
    ["javascript"] = javascript,
    ["javascriptreact"] = javascript,
    ["javascript.jsx"] = javascript,
    ["typescript"] = javascript,
    ["typescript.tsx"] = javascript,
    ["typescriptreact"] = javascript,
}

local on_buf_enter_cb = function()
    vim.opt_local.spell = false
    local filetype = vim.bo.filetype or "default"
    if (handlers[filetype]) then
        handlers[vim.bo.filetype]()
    end
end

function module.apply()
    local group = vim.api.nvim_create_augroup("OnBufEnter", {clear = true})
    vim.api.nvim_create_autocmd(
        "BufEnter", {group = group, pattern = "*", callback = on_buf_enter_cb}
    )
end

return module
