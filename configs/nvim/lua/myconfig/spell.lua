local module = {}

local present = require("myconfig.present");

local function spell_file_rebuild()
    local spell_dir = vim.env.HOME .. "/.dotfiles-private/nvim/spell"
    for _, spell_file_name in pairs(vim.fn.readdir(spell_dir)) do
        if (spell_file_name == string.gsub(spell_file_name, ".spl$", "")) then
            local spell_file_path = spell_dir .. "/" .. spell_file_name
            vim.cmd.mkspell {args = {spell_file_path}, bang = true}
        end
    end
end

local common = vim.env.HOME .. "/.dotfiles-private/nvim/spell/common.utf-8.add"
local natural = vim.env.HOME .. "/.dotfiles-private/nvim/spell/natural.utf-8.add"

function module.preset(name)
    if present.is_presenting() then
        return
    end
    local lang = vim.env.HOME .. "/.dotfiles-private/nvim/spell/" .. name .. ".utf-8.add"

    vim.opt_local.spell = true
    vim.opt_local.spellfile = common .. "," .. lang .. "," .. natural
end

function module.generic_preset()
    if present.is_presenting() then
        return
    end
    vim.opt_local.spell = true
    vim.opt_local.spellfile = common .. "," .. natural
end

function module.apply()
    vim.api.nvim_create_user_command("SpellFileRebuild", spell_file_rebuild, {nargs = 0})
end

function module.lsp_config()
    local dictionary = {}
    if vim.fn.filereadable(common) == 1 then
        for word in io.open(common, "r"):lines() do
            local trimed = vim.trim(word)
            if trimed ~= "" then
                table.insert(dictionary, word)
            end
        end
    end
    return {
        filetypes = {
            "bib", "gitcommit", "markdown", "org", "plaintex", "rst", "rnoweb", "tex", "pandoc",
            "quarto", "rmd", "context",
            -- "html",
            "xhtml"
        },
        settings = {
            ltex = {
                language = "en-US",
                additionalRules = {
                    -- https://languagetool.org/download/ngram-data/ngrams-en-20150817.zip
                    -- unzip to ~/.ngrams/en
                    languageModel = "~/.ngrams",
                },
                dictionary = {["en-US"] = dictionary},
                enabled = {
                    "bibtex", "gitcommit", "markdown", "org", "tex", "restructuredtext", "rsweave",
                    "latex", "quarto", "rmd", "context", "html", "xhtml"}
            },
        },
    }
end

return module
