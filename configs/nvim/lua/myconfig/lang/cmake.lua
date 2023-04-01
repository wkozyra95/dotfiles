local module = {}

local workspaces = require("myconfig.workspaces")

function module.cmake_config()
    if workspaces.current.vim.cmake_efm then
        return {
            on_attach = function(client)
                client.server_capabilities.documentFormattingProvider = false;
            end,
        }
    else
        return {}
    end
end

function module.attach_efm(config)
    if workspaces.current.vim.cmake_efm then
        config.settings.languages = vim.tbl_extend(
            "force", config.settings.languages, {cmake = {workspaces.current.vim.cmake_efm}}
        )
        config.filetypes = vim.list_extend(config.filetypes, {"cmake"})
        config.root_dir_patterns = vim.list_extend(config.root_dir_patterns, {"CMakeLists.txt"});
    end
end

return module
