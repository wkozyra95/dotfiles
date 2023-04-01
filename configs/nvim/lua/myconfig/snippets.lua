local module = {}

local ls = require("luasnip")

function module.apply()
    ls.config.set_config {
        history = true,
        updateevents = "TextChanged,TextChangedI",
        enable_autosnippets = true,
    }
end

local snippets_loaded = {}

function module.load(filetype, snippets)
    if snippets_loaded[filetype] then
        return
    end
    snippets_loaded[filetype] = true
    local result = {}
    for k, v in pairs(snippets) do
        table.insert(result, (ls.s({trig = k, desc = v.desc}, v)))
    end
    ls.add_snippets(filetype, result)
end

function module.expand_or_jump()
    if ls.expand_or_jumpable() then
        ls.expand_or_jump()
    end
end

function module.jump_back()
    if ls.jumpable(-1) then
        ls.jump(-1)
    end
end

function module.reload()
    snippets_loaded = {}
    package.loaded["myconfig.lang.go_snippets"] = nil
    module.load("go", require("myconfig.lang.go_snippets"))
    vim.cmd.edit()
end

return module
