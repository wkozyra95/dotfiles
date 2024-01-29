local _ = require("myconfig.utils")

local module = {}

function module.apply()
    vim.api.nvim_create_user_command(
        "ElixirLspInstall", function() _.rpc_run({name = "elixir:lsp:install"}) end, {nargs = 0}
    )
end

function module.elixirls_config() return {cmd = {"elixir-ls"}} end

return module;
