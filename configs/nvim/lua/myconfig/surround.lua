local _ = require("myconfig.utils")
local module = {}

--- @class Position
--- @field line integer
--- @field column integer
local Position = {}

--- @param line integer | nil
--- @param column integer | nil
--- @return Position | nil
function Position:new(line, column)
    local p = {line = line, column = column};
    if not line or not column then
        return nil
    end
    setmetatable(p, self)
    self.__index = self
    return p
end

--- @return Position
function Position:cursor()
    local current = vim.fn.getcurpos() -- 1 based indexed
    return assert(Position:new(current[2], current[3]))
end

function Position:is_before(position)
    assert(position)
    if self.line < position.line then
        return true;
    elseif self.line == position.line and self.column < position.column then
        return true;
    end
    return false;
end

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
--
--- @param position Position position where to start a search
--- @param symbol string symbol to search for
--- @return Position | nil
local function find_before(position, symbol)
    local lines = vim.api.nvim_buf_get_lines(
        0, _.max(position.line - MAX_SEARCH_CONTEXT, 0), position.line, false
    )
    local current_line = string.reverse(string.sub(lines[#lines], 0, position.column));
    local index_in_first_line = string.find(current_line, symbol, 1, true)
    if index_in_first_line then
        return Position:new(position.line, position.column - index_in_first_line + 1)
    end
    for i = #lines - 1, 1, -1 do
        local current_line_str = string.reverse(lines[i]);
        local reversed_index = string.find(current_line_str, symbol, 1, true)
        if reversed_index then
            return Position:new(position.line - #lines + i, #current_line_str - reversed_index + 1)
        end
    end
end

--- @param position Position
--- @param open_symbol string
--- @param close_symbol string
local function find_opening_bracket(position, open_symbol, close_symbol)
    local counter = 0;
    local last = position;
    while true do
        assert(counter >= 0)

        local last_open = find_before(last, open_symbol);
        if not last_open then
            return nil
        end
        local last_close = find_before(last, close_symbol);

        if counter == 0 then
            if (not last_close) or (last_close and not last_open:is_before(last_close)) then
                return last_open;
            end
        end
        if last_close then
            if last_open:is_before(last_close) then
                counter = counter + 1
                last = assert(Position:new(last_close.line, last_close.column - 1))
            else
                counter = counter - 1;
                last = assert(Position:new(last_open.line, last_open.column - 1))
            end
        else
            counter = counter - 1
            last = assert(Position:new(last_open.line, last_open.column - 1))
        end
    end
end

-- 1 based indexing for args and results
-- exclude symbol under the cursor
--
--- @param position Position position where to start a search
--- @param symbol string symbol to search for
--- @return Position | nil
local function find_after(position, symbol)
    local lines = vim.api.nvim_buf_get_lines(0, position.line - 1, position.line + MAX_SEARCH_CONTEXT,
        false) -- 0 based indexed args
    local current_line = string.sub(lines[1], position.column + 1);
    local index_in_first_line = string.find(current_line, symbol, 1, true)
    if index_in_first_line then
        return Position:new(position.line, position.column + index_in_first_line)
    end
    for i = 2, #lines, 1 do
        local current_line_str = lines[i];
        local index = string.find(current_line_str, symbol, 1, true)
        if index then
            return Position:new(position.line + (i - 1), index)
        end
    end
end

--- @param position Position
--- @param open_symbol string
--- @param close_symbol string
local function find_closing_bracket(position, open_symbol, close_symbol)
    local counter = 0;
    local last = position;
    while true do
        assert(counter >= 0)

        local first_close = find_after(last, close_symbol);
        if not first_close then
            return nil
        end
        local first_open = find_after(last, open_symbol);

        if counter == 0 then
            if (not first_open) or (first_open and first_close:is_before(first_open)) then
                return first_close;
            end
        end
        if first_open then
            if first_close:is_before(first_open) then
                counter = counter - 1
                last = assert(Position:new(first_close.line, first_close.column))
            else
                counter = counter + 1;
                last = assert(Position:new(first_open.line, first_open.column))
            end
        else
            counter = counter - 1
            last = assert(Position:new(first_close.line, first_close.column))
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
    local cursor_position = Position:cursor()
    local open_symbol, close_symbol = source, find_matchig_r_paren(source)
    local open = find_opening_bracket(cursor_position, open_symbol, close_symbol)
    local close = find_closing_bracket(cursor_position, open_symbol, close_symbol)
    if not open or not close then
        return
    end

    -- 0 based indexed args
    vim.api.nvim_buf_set_text(
        0, open.line - 1, open.column - 1, open.line - 1, open.column, {destination}
    )
    vim.api.nvim_buf_set_text(
        0, close.line - 1, close.column - 1, close.line - 1, close.column,
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
    local cursor_position = Position:cursor()
    local open_symbol, close_symbol = source, find_matchig_r_paren(source)
    local open = find_opening_bracket(cursor_position, open_symbol, close_symbol)
    local close = find_closing_bracket(cursor_position, open_symbol, close_symbol)
    P({open, close})
    if not open or not close then
        return
    end
    -- 0 based indexed args
    vim.api.nvim_buf_set_text(0, close.line - 1, close.column - 1, close.line - 1, close.column, {})
    vim.api.nvim_buf_set_text(0, open.line - 1, open.column - 1, open.line - 1, open.column, {})
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
