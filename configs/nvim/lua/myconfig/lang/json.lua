local workspace = require("myconfig.workspaces")

local module = {}

function module.jsonls_config()
    local schemas = {
        {
            fileMatch = {"app.json", "app.config.json"},
            url =
            "https://raw.githubusercontent.com/expo/vscode-expo/schemas/schema/expo-xdl.json",
        },
        {
            fileMatch = {"eas.json"},
            url =
            "https://raw.githubusercontent.com/expo/vscode-expo/schemas/schema/eas.json",
        },
        {
            fileMatch = {"store.config.json"},
            url =
            "https://raw.githubusercontent.com/expo/vscode-expo/schemas/schema/eas-metadata.json",
        },
    }
    if workspace.current and workspace.current.vim.json_schemas then
        schemas = vim.list_extend(schemas, workspace.current.vim.json_schemas)
    end
    return {
        settings = {
            json = {
                schemas = schemas,
                validate = {enable = true},
            },
        },
    }
end

return module
