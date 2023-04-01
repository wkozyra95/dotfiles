local debug = require("myconfig.debug")
local utils = require("myconfig.utils")
local tel = require("myconfig.telescope")
local tel_builtin = require("telescope.builtin")
local tree = require("myconfig.treesitter")
local lsp = require("myconfig.lsp")
local workspaces = require("myconfig.workspaces")
local actions = require("myconfig.actions")
local snippets = require("myconfig.snippets")
local neogit = require("neogit")
local git = require("myconfig.git")
local surround = require("myconfig.surround")

local ref = utils.ref
local remap = {noremap = false}

-- # Modes:
--   - normal + select + visual + operator pending
-- n - normal mode
-- v - visual + select
-- s - select (<c-g> from visual to enter, behaves like normal editor)
-- x - visual
-- o - operator pending (in the middle of keybinding unless no wait)
-- i - insert (+replace)
-- l - insert + commandline + lang
-- c - commandline (:somecommand)
-- t - terminal (can be launched with :terminal)
--
-- # Options:
-- noremap - if rhs is binding use original behavior (if noremap is false use remapped behavior)
-- silent - ignore output from commands
-- buffer - current buffer only
-- nowait - execute keybinding even if there is other binding that could match with next keystroke

local mapping = {
    ["<space>"] = {
        {"<space><c-g>u", "i"}, -- history breakpoint <c-g>u
        ["<space>"] = {tel.file_browser_current_dir},
        a = {
            a = {lsp.codeAction},
            r = {lsp.rename},
            q = {lsp.autoFix},
            f = {{lsp.format, "n"}, {lsp.formatSelected, "v"}},
        },
        f = {
            f = {actions.actions.grep_files.fn},
            g = {c = {tel_builtin.git_commits}, b = {tel_builtin.git_branches}},
            v = {actions.actions.dotfiles_search.fn},
            n = {actions.actions.notes_search.fn},
            h = {tel_builtin.help_tags},
            b = {tel_builtin.buffers},
            r = {tel_builtin.resume},
        },
        c = {d = {lsp.showLineDiagnostics}},
        s = {s = {workspaces.switch_workspace}, d = {workspaces.switch_to_current_dir}},
        g = {
            s = {function() vim.cmd.vertical("G") end},
            b = {function() vim.cmd.vertical("G blame") end},
            g = {neogit.open},
            d = {git.toggle_diffview},
            h = {git.toggle_diffhistory},
        },
        d = {
            l = {lsp.restart},
            p = {debug.playground, "", {silent = false}},
            r = {debug.reload},
            s = {snippets.reload},
        },
        ["<bs>"] = {":<Up>", "", {silent = false}},
    },
    [",,"] = {actions.select_action(), ""},
    ["<C-"] = {
        ["k>"] = {snippets.expand_or_jump, {"i", "s"}},
        ["j>"] = {snippets.jump_back, {"i", "s"}},
        ["e>"] = {tel.file_browser_root},
        ["n>"] = {"<cmd>tab split<cr>"},
        ["p>"] = {actions.actions.find_files.fn},
        ["s>"] = {
            ref(tree, "selection_init"),
            ref(tree, "selection_inc"),            -- active after selection init
        },
        ["a>"] = ref(tree, "selection_inc_scope"), -- active after selection init
        ["x>"] = ref(tree, "selection_dec"),       -- active after selection init
        ["h>"] = {lsp.onHover},
    },
    ["<bs>"] = {"<C-^>", "n"}, -- switch alternative buffer
    ["?"] = {tel_builtin.current_buffer_fuzzy_find},
    H = {"gT"},
    L = {"gt"},
    Y = {"y$", "n"},
    Q = {"<nop>", "", remap},
    g = {
        -- g = {} - already used,
        b = {{ref(tree, "comment_block")}, b = {ref(tree, "comment_toggle_block")}},
        c = {{ref(tree, "comment_line")}, c = {ref(tree, "comment_toggle_line")}},
        d = {lsp.goToDefinition},
        D = {lsp.goToDeclaration},
        t = {lsp.goToTypeDefinition},
        r = {lsp.references},
        i = {lsp.goToPrev},
        u = {lsp.goToNext},
    },
    d = {
        -- d = {} - already used,
        s = {surround.remove, "n"},
    },
    c = {
        -- c = {} - already used,
        s = {surround.replace, "n"},
    },
    t = {t = {"<cmd>terminal<cr>", "n"}},
    ["\""] = {surround.surround_selection({"\""}, {"\""}), "x"},
    ["'"] = {surround.surround_selection({"'"}, {"'"}), "x"},
    ["`"] = {
        {surround.surround_selection({"`"}, {"`"}), "x"},
        {surround.auto_pair({"`"}, {"`"}),          "i"},
    },
    ["{"] = {
        {surround.surround_selection({"{"}, {"}"}), "x"},
        {surround.auto_pair({"{"}, {"}"}),          "i"},
    },
    ["("] = {
        {surround.surround_selection({"("}, {")"}), "x"},
        {surround.auto_pair({"("}, {")"}),          "i"},
    },
    ["["] = {
        {surround.surround_selection({"["}, {"]"}), "x"},
        {surround.auto_pair({"["}, {"]"}),          "i"},
    },
    ["<"] = {surround.surround_selection({"<"}, {">"}), "x"},
    ["}"] = {surround.surround_selection({"{", ""}, {"", "}"}), "x"},
    [")"] = {surround.surround_selection({"(", ""}, {"", ")"}), "x"},
    ["]"] = {surround.surround_selection({"[", ""}, {"", "]"}), "x"},
    --
    -- biddings keep the same general purpose, just side effects
    --
    {
        -- history breakpoint <c-g>u
        [","] = {",<c-g>u", "i"},
        ["."] = {".<c-g>u", "i"},
        -- do not update copy register, send instead to black hole register
        c = {"\"_c", ""},
        C = {"\"_C", ""},
        x = {"\"_x", ""},
        X = {"\"_X", ""},
        -- do not replace copy register when pasting in visual mode
        p = {"\"_dP", "v"},
    },
}

local default_options = {noremap = true, silent = true}
utils.apply_keymap(mapping, default_options)
