local module = {}

local node_stack = {}

local esc = vim.api.nvim_replace_termcodes("<Esc>", true, false, true)

local expanding = false

local function select_node(node)
    local sr, sc, er, ec = node:range()
    local end_line, end_col
    if ec == 0 and er > 0 then
        end_line = er
        end_col = math.max(vim.fn.col({end_line, "$"}) - 2, 0)
    else
        end_line = er + 1
        end_col = math.max(ec - 1, 0)
    end

    expanding = true
    if vim.fn.mode():lower():match("[v\22]") then
        vim.api.nvim_feedkeys(esc, "nx", false)
    end

    vim.api.nvim_win_set_cursor(0, {sr + 1, sc})
    vim.cmd("normal! v")
    vim.api.nvim_win_set_cursor(0, {end_line, end_col})
    vim.schedule(function() expanding = false end)
end

local function seed_node()
    local row = vim.fn.line(".") - 1
    local col = vim.fn.col(".") - 1
    local line_text = vim.fn.getline(row + 1)
    local first_nonblank = line_text:find("[^%s]")
    if first_nonblank and col < first_nonblank - 1 then
        col = first_nonblank - 1
    end
    local ok, parser = pcall(vim.treesitter.get_parser, 0)
    if not ok or not parser then return nil end
    local tree = parser:parse()[1]
    if not tree then return nil end
    return tree:root():descendant_for_range(row, col, row, col)
end

local function range_eq(a, b)
    local asr, asc, aer, aec = a:range()
    local bsr, bsc, ber, bec = b:range()
    return asr == bsr and asc == bsc and aer == ber and aec == bec
end

function module.expand_node()
    local buf = vim.api.nvim_get_current_buf()
    local stack = node_stack[buf] or {}

    local prev, node
    if #stack == 0 then
        local seed = seed_node()
        if not seed then return end
        table.insert(stack, seed)
        prev = seed
        node = seed:parent()
    else
        prev = stack[#stack]
        node = prev:parent()
    end
    while node and range_eq(node, prev) do
        node = node:parent()
    end
    if not node then return end

    table.insert(stack, node)
    node_stack[buf] = stack
    select_node(node)
end

function module.start_selection()
    local row = vim.fn.line(".")
    local col = vim.fn.col(".")
    local line_text = vim.fn.getline(row)
    local first_nonblank = line_text:find("[^%s]")
    if first_nonblank and col < first_nonblank then
        vim.api.nvim_win_set_cursor(0, {row, first_nonblank - 1})
    end
    vim.cmd("normal! viw")
end

function module.shrink_node()
    local buf = vim.api.nvim_get_current_buf()
    local stack = node_stack[buf] or {}
    if #stack <= 1 then return end
    table.remove(stack)
    node_stack[buf] = stack
    select_node(stack[#stack])
end

function module.apply()
    vim.api.nvim_create_autocmd("FileType", {
        callback = function(args)
            local lang = vim.treesitter.language.get_lang(args.match)
            if not lang then return end
            pcall(vim.treesitter.start, args.buf, lang)
        end,
    })

    vim.api.nvim_create_autocmd("BufWipeout", {
        callback = function(args) node_stack[args.buf] = nil end,
    })

    vim.api.nvim_create_autocmd({"CursorMoved", "ModeChanged"}, {
        callback = function(args)
            if expanding then return end
            node_stack[args.buf] = nil
        end,
    })

    require("treesitter-context").setup {
        enable = true,
        throttle = true,
        max_lines = 0,
        patterns = {
            default = {"class", "function", "method", "for", "while", "if", "switch", "case"},
            json = {"object", "array"},
            yaml = {"block_mapping_pair"},
        },
        exact_patterns = {}
    }
end

return module
