local workspaces = require("myconfig.workspaces")

local module = {}

function module.tsserver_config()
    return {
        init_options = {
            hostInfo = "neovim",
            preferences = {importModuleSpecifierPreference = "non-relative"},
        },
        on_attach = function(client)
            if workspaces.current.vim.eslint then
                client.server_capabilities.documentFormattingProvider = false
            end
        end,
        filetypes = {
            "javascript",
            "javascriptreact",
            "javascript.jsx",
            "typescript",
            "typescript.tsx",
            "typescriptreact",
        },
    }
end

function module.eslint_config()
    return {
        on_attach = function(client)
            if workspaces.current.vim.eslint then
                client.server_capabilities.documentFormattingProvider = true
            end
        end,
        settings = {packageManager = "yarn"},
    }
end

return module;
