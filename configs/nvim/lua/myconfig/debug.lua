local module = {}

module.enabled = true

function module.reload()
    for k, _ in pairs(package.loaded) do
        if vim.startswith(k, "myconfig") then
            print("reload " .. k)
            package.loaded[k] = nil
        end
    end
    vim.cmd.luafile({args = {"/home/wojtek/.config/nvim/init.lua"}})
    print("reloaded");
end

function module.playground()
    vim.cmd.messages("clear")
end

function module.apply()
end

return module;
