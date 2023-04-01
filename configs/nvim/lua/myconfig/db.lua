local module = {}

local workspace = require("myconfig.workspaces")

function module.apply()
    vim.g.dbs = workspace.current.vim.databases or {}
    vim.g.db_ui_tmp_query_location = "~/.cache/db"
    vim.g.db_ui_execute_on_save = 0 -- disabled
    vim.g.db_ui_use_nerd_fonts = 1
    vim.g.db_ui_force_echo_notifications = 1
end

return module;
