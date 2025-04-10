local _ = require("myconfig.utils")

local module = {}

local path = {
    data = function() return vim.fn.stdpath("data") end, -- "~/.local/share/nvim"
}

local sumneko_command = function()
    local bin = "lua-language-server"
    return {bin}
end

local function get_runtime()
    local result = {};
    -- local plugin_path = path.data() .. "/lazy"
    -- for _, plugin in pairs(vim.fn.readdir(plugin_path)) do
    --     local lua_path = plugin_path .. "/" .. plugin .. "/lua";
    --     if vim.fn.isdirectory(lua_path) == 1 then
    --         result[lua_path] = true
    --     end
    -- end
    result[vim.fn.expand("$VIMRUNTIME/lua")] = true
    result[vim.fn.expand("$VIMRUNTIME/lua/vim/lsp")] = true
    return result
end

function module.lua_ls_config()
    return {
        cmd = sumneko_command(),
        settings = {
            Lua = {
                runtime = {version = "LuaJIT"},
                completion = {keywordSnippet = "Disable"},
                diagnostics = {
                    enable = true,
                    disable = {"trailing-space"},
                    neededFileStatus = {["codestyle-check"] = "Any"},
                    globals = {
                        -- Neovim
                        "vim",
                        -- Busted
                        "describe",
                        "it",
                        "before_each",
                        "after_each",
                        "teardown",
                        "pending",
                        "clear",
                    },
                },
                format = {
                    enable = true,
                    defaultConfig = {
                        indent_style = "space",
                        indent_size = "4",
                        quote_style = "double",
                        max_line_length = "101",
                        space_around_table_field_list = "false",
                    },
                },
                workspace = {
                    library = get_runtime(),
                    maxPreload = 10000,
                    preloadFileSize = 10000,
                    checkThirdParty = false,
                },
            },
        },
        filetypes = {"lua"},
        handlers = {},
    }
end

return module;
