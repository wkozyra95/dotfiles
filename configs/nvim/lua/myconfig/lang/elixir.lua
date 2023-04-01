local _ = require("myconfig.utils")

local module = {}

local path = {
    language_server_script = function()
        return vim.fn.expand("~") .. "/.cache/nvim/myconfig/elixirls/language_server.sh"
    end,
}

function module.apply()
    vim.api.nvim_create_user_command(
        "ElixirLspInstall", function() _.rpc_run({name = "elixir:lsp:install"}) end, {nargs = 0}
    )
end

function module.elixirls_config() return {cmd = {path.language_server_script()}} end

return module;
