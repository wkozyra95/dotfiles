local module = {}

local _ = require("myconfig.utils")

local path = {
    cache = function() return vim.fn.stdpath("cache") end, -- "~/.cache/nvim"
}

local playground_buffers = {}

local function start_playground(user_command_opts)
    local fargs = user_command_opts.fargs or {}
    local name = fargs[1]
    module.ensure_commands()
    local playground_path = name and (path.cache() .. "/node/playground/" .. name) or
        ("/tmp/nvim_node_" .. vim.fn.localtime())
    _.rpc_run({name = "node:playground:create", path = playground_path})
    local buffers = vim.api.nvim_list_bufs()
    for _, b in ipairs(buffers) do
        local bufname = vim.api.nvim_buf_get_name(b)
        if bufname ~= "" and string.sub(bufname, 0, 8) ~= "noice://" then
            vim.api.nvim_buf_delete(b, {})
        end
    end
    vim.fn.chdir(playground_path)
    vim.cmd.edit(playground_path .. "/index.ts")

    local buffer = vim.api.nvim_get_current_buf()
    playground_buffers[buffer] = {path = playground_path}
    vim.api.nvim_buf_attach(
        buffer, false, {
            on_detach = function()
                if not name then
                    -- _.rpc_run({name = "node:playground:delete", path = playground_path})
                    playground_buffers[buffer] = nil
                end
            end,
        }
    );
end

local function install_package(user_command_opts)
    local fargs = user_command_opts.fargs or {}
    local package = fargs[1]
    local buffer = vim.api.nvim_get_current_buf()
    local playground_path = playground_buffers[buffer] and playground_buffers[buffer].path
    if playground_path then
        _.rpc_run({name = "node:playground:install", path = playground_path, package = package})
        vim.cmd.edit({bang = true})
    else
        vim.notify("Install supported only in playground buffers")
    end
end

local function playground_zsh_shell()
    local buffer = vim.api.nvim_get_current_buf()
    local playground_path = playground_buffers[buffer] and playground_buffers[buffer].path
    if playground_path then
        _.rpc_start({name = "node:playground:zsh-shell", path = playground_path}, function()
        end)
    end
end

local function playground_node_shell()
    local buffer = vim.api.nvim_get_current_buf()
    local playground_path = playground_buffers[buffer] and playground_buffers[buffer].path
    if playground_path then
        _.rpc_start({name = "node:playground:node-shell", path = playground_path}, function()
        end)
    end
end

local function playground_complete()
    local files = _.rpc_run({name = "directory:preview", path = path.cache() .. "/node/playground"})
    local names = {}
    for _, file in ipairs(files) do
        table.insert(names, file.name)
    end
    return names
end

module.ensure_commands = _.once(
    function()
        vim.api
            .nvim_create_user_command("NodePlaygroundNodeShell", playground_node_shell, {nargs = 0})
        vim.api
            .nvim_create_user_command("NodePlaygroundZshShell", playground_zsh_shell, {nargs = 0})
        vim.api
            .nvim_create_user_command("NodePlaygroundInstallPackage", install_package, {nargs = 1})
    end
)

module.apply = function()
    vim.api.nvim_create_user_command(
        "NodePlayground", start_playground, {nargs = "?", complete = _.once(playground_complete)}
    )
end

return module
