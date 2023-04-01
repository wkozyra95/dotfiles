local module = {}

local _ = require("myconfig.utils")

local path = {
    cache = function() return vim.fn.stdpath("cache") end, -- "~/.cache/nvim"
}

local playground_buffers = {}

local function start_playground(user_command_opts)
    local fargs = user_command_opts.fargs or {}
    local image = fargs[1]
    module.ensure_commands()
    local playground_path = image and (path.cache() .. "/playground/" .. image) or
        ("/tmp/nvim_docker_" .. vim.fn.localtime())
    _.rpc_run({name = "docker:playground:create", path = playground_path, image = image or "ubuntu"})
    local buffers = vim.api.nvim_list_bufs()
    for _, b in ipairs(buffers) do
        local bufname = vim.api.nvim_buf_get_name(b)
        if bufname ~= "" and string.sub(bufname, 0, 8) ~= "noice://" then
            vim.api.nvim_buf_delete(b, {})
        end
    end
    vim.cmd.cd(playground_path)
    vim.cmd.edit(playground_path .. "/Dockerfile")

    local buffer = vim.api.nvim_get_current_buf()
    playground_buffers[buffer] = {path = playground_path}
    vim.api.nvim_buf_attach(
        buffer, false, {
            on_detach = function()
                if not image then
                    -- _.rpc_run({name = "docker:playground:delete", path = playground_path})
                    playground_buffers[buffer] = nil
                end
            end,
        }
    );
end

local function playground_shell()
    local buffer = vim.api.nvim_get_current_buf()
    local playground_path = playground_buffers[buffer] and playground_buffers[buffer].path
    if playground_path then
        _.rpc_start({name = "docker:playground:shell", path = playground_path}, function()
        end)
    end
end

local function playground_complete()
    local files = _.rpc_run({name = "directory:preview", path = path.cache() .. "/playground"})
    local names = {}
    for _, file in ipairs(files) do
        table.insert(names, file.name)
    end
    return names
end

module.ensure_commands = _.once(
    function()
        vim.api.nvim_create_user_command("DockerPlaygroundShell", playground_shell, {nargs = 0})
    end
)

module.apply = function()
    vim.api.nvim_create_user_command(
        "DockerPlayground", start_playground, {nargs = "?", complete = _.once(playground_complete)}
    )
end

return module
