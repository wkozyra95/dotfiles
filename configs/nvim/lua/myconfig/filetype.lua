local module = {}

local format = require("myconfig.format")
local spell = require("myconfig.spell")
local snippets = require("myconfig.snippets")
local workspace = require("myconfig.workspaces")
local present = require("myconfig.present")

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
    elixir = function()
        require("myconfig.lang.elixir").apply()
        spell.preset("elixir")
    end,
    json = format.preset_2,
    lua = function()
        format.preset(4)
        spell.preset("lua")
        snippets.load("lua", require("myconfig.lang.lua_snippets"))
    end,
    markdown = function()
        format.preset(4)
        spell.generic_preset()
    end,
    gitcommit = spell.strict_preset,
    plist = format.preset_4,
    yaml = format.preset_2,
    nix = format.preset_2,
    ["javascript"] = javascript,
    ["javascriptreact"] = javascript,
    ["javascript.jsx"] = javascript,
    ["typescript"] = javascript,
    ["typescript.tsx"] = javascript,
    ["typescriptreact"] = javascript,
}

local apply_workspace_settings = function(filetype)
    if (not workspace.current.name) then
        return
    end
    local config = (workspace.current.vim.filetype_config or {})[filetype];
    if (config) then
        if (config.indent_size and config.indent_size > 0) then
            format.preset(config.indent_size)
        end
    end
end

local on_buf_enter_cb = function()
    vim.opt_local.spell = false
    local filetype = vim.bo.filetype or "default"
    if (handlers[filetype]) then
        handlers[filetype]()
    end
    apply_workspace_settings(filetype)
    present.on_buf_enter_hook()
end

function module.apply()
    local group = vim.api.nvim_create_augroup("OnBufEnter", {clear = true})
    vim.api.nvim_create_autocmd(
        "BufEnter", {group = group, pattern = "*", callback = on_buf_enter_cb}
    )
end

return module
