local module = {}

local workspaces = require("myconfig.workspaces")
local parsers = require("nvim-treesitter.parsers")

function module.gopls_config()
    if workspaces.current.vim.go_efm then
        return {
            on_attach = function(client)
                client.server_capabilities.documentFormattingProvider = false;
            end,
            settings = {gopls = {buildFlags = {"-tags=e2e"}}},
        }
    else
        return {
            settings = {gopls = {buildFlags = {"-tags=e2e"}}},
        }
    end
end

function module.golangci_config()
    return {
        on_attach = function(client)
            client.server_capabilities.documentFormattingProvider = false;
        end,
    }
end

function module.attach_efm(config)
    if workspaces.current.vim.go_efm then
        config.settings.languages = vim.tbl_extend(
            "force", config.settings.languages, {go = {workspaces.current.vim.go_efm}}
        )
        config.filetypes = vim.list_extend(config.filetypes, {"go"})
        config.root_dir_patterns = vim.list_extend(config.root_dir_patterns, {"go.sum", "go.mod"});
    end
end

local function get_start_of_next_sibling(node)
    local next = node:next_named_sibling()
    if next == nil then
        local _, _, end_row, _ = node:range()
        return end_row, -1
    end
    local start_row, start_col, _, _ = next:range()
    return start_row, start_col
end

local function update_node(node)
    local _, _, end_row, end_col = node:range()
    local lines = vim.api.nvim_buf_get_lines(0, end_row, end_row + 1, false)
    local next_start_row, next_start_col = get_start_of_next_sibling(node)
    local suffix_end_col = next_start_row == end_row and next_start_col or -1
    local suffix = lines[1]:sub(end_col + 1, suffix_end_col)
    local match = suffix:match(" *$")
    if match == suffix then
        local lines_slice = {lines[1]:sub(0, end_col) .. "," .. lines[1]:sub(end_col + 1, -1)}
        vim.api.nvim_buf_set_lines(0, end_row, end_row + 1, false, lines_slice)
    end
end

local function get_last_non_comment(last_node)
    local node = last_node
    while node and node:type() == "comment" do
        node = node:prev_named_sibling()
    end
    return node
end

function module.format(original_format)
    local tree = parsers.get_parser(0):parse()[1]
    local query = vim.treesitter.query.get("go", "trailing_commas")
    if not query then
        vim.notify("Treesitter query \"trailing_commas\" not found for \"go\" filetype.")
        return
    end
    for _, node in query:iter_captures(tree:root(), 0, 0, -1) do
        local non_comment = get_last_non_comment(node)
        if non_comment then
            update_node(non_comment)
        end
    end
    if original_format then
        original_format()
    end
end

return module
