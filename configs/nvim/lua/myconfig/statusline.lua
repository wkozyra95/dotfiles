local Job = require("plenary.job")

local green = "#b8bb26"
local red = "#fb4934"
local blue = "#83a598"
local yellow = "#fabd2f"
local purple = "#d3869b"
local black = "#504945"
local white = "#bdae93"

local icons = {
    indicator_errors = "",
    indicator_warnings = "",
    indicator_ok = "",
    indicator_load = "",
}

local modeConfigs = {
    ["n"] = {name = "N", fg = green},
    ["v"] = {name = "V", fg = purple},
    ["V"] = {name = "V·L", fg = purple},
    ["\22"] = {name = "V·B", fg = purple},
    ["i"] = {name = "I", fg = blue},
    ["R"] = {name = "R", fg = red},
    ["Rv"] = {name = "V·R", fg = red},
    ["c"] = {name = "C", fg = yellow},
}

-- local left_separator = ""
local right_separator = ""

local function set_mode_hl(modeConfig)
    vim.api.nvim_set_hl(
        0, "StatusLineModeSeparator", {
            foreground = modeConfig.fg,
            background = white,
            reverse = false,
            underline = false,
            bold = false,
        }
    )
    vim.api.nvim_set_hl(
        0, "StatusLineMode", {
            foreground = modeConfig.fg,
            background = black,
            reverse = true,
            underline = false,
            bold = true,
        }
    )
end

vim.schedule(
    function()
        vim.api.nvim_set_hl(
            0, "StatusLineGitSeparator", {
                foreground = white,
                background = black,
                reverse = false,
                underline = false,
                bold = true,
            }
        )
        vim.api.nvim_set_hl(
            0, "StatusLineGit", {
                foreground = white,
                background = black,
                reverse = true,
                underline = false,
                bold = true,
            }
        )
    end
)

local function git_branch()
    local j = Job:new(
    ---@diagnostic disable-next-line: missing-fields
        {command = "git", args = {"branch", "--show-current"}, cwd = vim.fn.fnamemodify("", ":h")}
    )

    local ok, result = pcall(function() return vim.trim(j:sync()[1]) end)

    if ok then
        return result
    end
end

local function lsp_status()
    local status = {}
    local errors = #vim.diagnostic.get(0, {severity = vim.diagnostic.severity.ERROR})
    if errors and errors > 0 then
        table.insert(status, icons.indicator_errors .. " " .. errors)
    end

    local warnings = #vim.diagnostic.get(0, {severity = vim.diagnostic.severity.WARN})
    if warnings and warnings > 0 then
        table.insert(status, icons.indicator_warnings .. " " .. warnings)
    end
    local progress = vim.lsp.status()
    if progress ~= "" then
        return progress .. " " .. icons.indicator_load .. " "
    end
    return #status > 0 and table.concat(status, " ") or icons.indicator_ok .. " "
end

local function statusline()
    -- item is defined as %-0{minwid}.{maxwid}{item}
    local rawMode = vim.api.nvim_get_mode().mode
    local modeConfig = modeConfigs[rawMode] or {name = rawMode}
    set_mode_hl(modeConfig)
    local mode = "%#StatusLineMode# " .. modeConfig.name .. " %#StatusLineModeSeparator#" ..
        right_separator .. "%#StatusLine#"
    local file = "%f"

    --local lsp = ""
    --if (#vim.lsp.get_clients() > 0) then
    --    lsp = lsp_status() .. " "
    --end

    -- for some reason it breaks noice.nvim
    -- local git = ""
    -- local git_current_branch = git_branch()
    -- if git_current_branch then
    --     git = "%#StatusLineGit# " .. git_current_branch .. "%#StatusLineGitSeparator#" ..
    --               right_separator .. "%#StatusLine#"
    -- end

    local position = "%#StatusLineGit# %2.p%% [%3.l/%L] %c"
    return mode .. file .. "%=" .. position
end

_G.statusline = statusline

vim.opt.statusline = "%!v:lua.statusline()"
vim.opt.winbar = "%#StatusLine#%=%m %f"
