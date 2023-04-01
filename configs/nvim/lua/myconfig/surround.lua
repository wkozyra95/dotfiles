local _ = require("myconfig.utils")
local module = {}

local function get_visual_pos()
    local start = vim.fn.getpos("v")
    local eend = vim.fn.getcurpos()
    local start_line = start[2] - 1
    local start_col = start[3] - 1
    local end_line = eend[2] - 1
    local end_col = eend[3] - 1
    if (start_line > end_line or (start_line == end_line and start_col > end_col)) then
        local end_line_length = #vim.fn.getline(start_line + 1)
        return end_line, end_col, start_line, math.min(start_col + 1, end_line_length)
    else
        local end_line_length = #vim.fn.getline(end_line + 1)
        return start_line, start_col, end_line, math.min(end_line_length, end_col + 1)
    end
end

function module.surround_selection(open_sign, close_sign)
    return function()
        local start_line, start_col, end_line, end_col = get_visual_pos();
        vim.api.nvim_buf_set_text(0, end_line, end_col, end_line, end_col, close_sign)
        vim.api.nvim_buf_set_text(0, start_line, start_col, start_line, start_col, open_sign)
        vim.api.nvim_feedkeys("=", "v", false)
        vim.defer_fn(
            function()
                -- TODO: very hacky, do it cleaner
                vim.api.nvim_win_set_cursor(0, {start_line + 1, start_col})
            end, 100
        )
    end
end

local valid_auto_pairs = {{"{", "}"}, {"(", ")"}, {"[", "]"}, {"`", "`"}, {"```", "```"}}
local valid_replecements = {
    {"{",  "}"},
    {"(",  ")"},
    {"[",  "]"},
    {"\"", "\""},
    {"'",  "'"},
    {"`",  "`"},
    {"<",  ">"},
}
local MAX_SEARCH_CONTEXT = 20

-- 1 based indexing for args and results
-- include symbol under the cursor
local function find_before(line, column, symbol)
    local lines = vim.api.nvim_buf_get_lines(0, _.max(line - MAX_SEARCH_CONTEXT, 0), line, false)
    local current_line = string.reverse(string.sub(lines[#lines], 0, column));
    local index_in_first_line = string.find(current_line, symbol, 1, true)
    if index_in_first_line then
        return line, column - index_in_first_line + 1
    end
    for i = #lines - 1, 1, -1 do
        local current_line_str = string.reverse(lines[i]);
        local reversed_index = string.find(current_line_str, symbol, 1, true)
        if reversed_index then
            return line - #lines + i, #current_line_str - reversed_index + 1
        end
    end
end

-- 1 based indexing for args and results
-- exclude symbol under the cursor
local function find_after(line, column, symbol)
    local lines = vim.api.nvim_buf_get_lines(0, line - 1, line + MAX_SEARCH_CONTEXT, false) -- 0 based indexed args
    local current_line = string.sub(lines[1], column + 1);
    local index_in_first_line = string.find(current_line, symbol, 1, true)
    if index_in_first_line then
        return line, column + index_in_first_line
    end
    for i = 2, #lines, 1 do
        local current_line_str = lines[i];
        local index = string.find(current_line_str, symbol, 1, true)
        if index then
            return line + (i - 1), index
        end
    end
end

local function find_matchig_r_paren(l_paren)
    for _, symbols in ipairs(valid_replecements) do
        if symbols[1] == l_paren then
            return symbols[2]
        end
    end
end

local function find_matchig_r_auto_pair(l_paren)
    for _, symbols in ipairs(valid_auto_pairs) do
        if symbols[1] == l_paren then
            return symbols[2]
        end
    end
end

local function is_valid_l_paren(l_paren)
    for _, symbols in ipairs(valid_replecements) do
        if symbols[1] == l_paren then
            return true
        end
    end
    return false
end

function module.replace()
    local source = vim.fn.getchar()
    if source == 27 then
        return
    end
    source = string.char(source)
    if not is_valid_l_paren(source) then
        return
    end
    local destination = vim.fn.getchar()
    if destination == 27 then
        return
    end
    destination = string.char(destination)
    if not is_valid_l_paren(destination) then
        return
    end
    local current = vim.fn.getcurpos() -- 1 based indexed
    local line = current[2]
    local col = current[3]
    local open_line, open_col = find_before(line, col, source)
    local close_line, close_col = find_after(line, col, find_matchig_r_paren(source))

    -- 0 based indexed args
    vim.api.nvim_buf_set_text(
        0, open_line - 1, open_col - 1, open_line - 1, open_col, {destination}
    )
    vim.api.nvim_buf_set_text(
        0, close_line - 1, close_col - 1, close_line - 1, close_col,
        {find_matchig_r_paren(destination)}
    )
end

function module.remove()
    local source = vim.fn.getchar()
    if source == 27 then
        return
    end
    source = string.char(source)
    if not is_valid_l_paren(source) then
        return
    end
    local current = vim.fn.getcurpos() -- 1 based indexed
    local line = current[2]
    local col = current[3]
    local open_line, open_col = find_before(line, col, source)
    local close_line, close_col = find_after(line, col, find_matchig_r_paren(source))
    -- 0 based indexed args
    vim.api.nvim_buf_set_text(0, close_line - 1, close_col - 1, close_line - 1, close_col, {})
    vim.api.nvim_buf_set_text(0, open_line - 1, open_col - 1, open_line - 1, open_col, {})
end

function module.auto_pair(start, eend)
    return function()
        local current = vim.fn.getcurpos()
        local line = current[2]
        local col = current[3]
        vim.api.nvim_buf_set_text(0, line - 1, col - 1, line - 1, col - 1, eend)
        vim.api.nvim_buf_set_text(0, line - 1, col - 1, line - 1, col - 1, start)
        vim.api.nvim_win_set_cursor(0, {line, col})
    end
end

return module
