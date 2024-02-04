local module = {}

local function spell_file_rebuild()
    local spell_dir = vim.env.HOME .. "/.dotfiles-private/nvim/spell"
    for _, spell_file_name in pairs(vim.fn.readdir(spell_dir)) do
        if (spell_file_name == string.gsub(spell_file_name, ".spl$", "")) then
            local spell_file_path = spell_dir .. "/" .. spell_file_name
            vim.cmd.mkspell {args = {spell_file_path}, bang = true}
        end
    end
end

function module.preset(name)
    local common = vim.env.HOME .. "/.dotfiles-private/nvim/spell/common.utf-8.add"
    local lang = vim.env.HOME .. "/.dotfiles-private/nvim/spell/" .. name .. ".utf-8.add"
    local natural = vim.env.HOME .. "/.dotfiles-private/nvim/spell/natural.utf-8.add"

    vim.opt_local.spell = true
    vim.opt_local.spellfile = common .. "," .. lang .. "," .. natural
end

function module.strict_preset()
    vim.opt_local.spell = true
    vim.opt_local.spellcapcheck = "[.?!]\\_[\\])'\" \\t]\\+"

    vim.opt_local.spellfile = vim.env.HOME .. "/.dotfiles-private/nvim/spell/natural.utf-8.add"
end

function module.apply()
    vim.api.nvim_create_user_command("SpellFileRebuild", spell_file_rebuild, {nargs = 0})
end

return module
