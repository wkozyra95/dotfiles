local module = {}

local function trim_whitespaces()
    local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false);
    for i, line in ipairs(lines) do
        lines[i] = string.gsub(line, "%s+$", "")
    end
    vim.api.nvim_buf_set_lines(0, 0, -1, true, lines)
end

function module.apply()
    vim.api.nvim_create_user_command("TrimWhitespaces", trim_whitespaces, {nargs = 0})
end

function module.preset(number)
    vim.opt_local.tabstop = number
    vim.opt_local.shiftwidth = number
    vim.opt_local.softtabstop = number
    vim.opt_local.expandtab = true
end

function module.preset_2() module.preset(2) end

function module.preset_4() module.preset(4) end

return module
