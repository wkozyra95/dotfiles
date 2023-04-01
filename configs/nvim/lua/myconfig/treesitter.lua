local module = {}

local ts = require("nvim-treesitter.configs")
local tsi = require("nvim-treesitter.incremental_selection")

module.incremental_selection_init = tsi.init_selection
module.selection_init = nil
module.selection_inc = nil
module.selection_dec = nil
module.selection_inc_scope = nil

module.comment_line = nil
module.comment_line = nil
module.comment_toggle_line = nil
module.comment_toggle_block = nil

local ensure_installed = {
    "bash",
    "bibtex",
    "c",
    "c_sharp",
    "clojure",
    "cmake",
    "comment",
    "cpp",
    "css",
    "dart",
    "dockerfile",
    "eex",
    "elixir",
    "erlang",
    "fish",
    "glsl",
    "go",
    "gomod",
    "gowork",
    "graphql",
    "hcl",
    "heex",
    "help",
    "hjson",
    "html",
    "java",
    "javascript",
    "jsdoc",
    "json",
    "json5",
    "jsonc",
    "kotlin",
    "latex",
    "llvm",
    "lua",
    "make",
    "ninja",
    "nix",
    "norg",
    "perl",
    "php",
    "python",
    "query",
    "regex",
    "ruby",
    "rust",
    "scala",
    "scss",
    "swift",
    "scheme",
    "teal",
    "toml",
    "tsx",
    "typescript",
    "vala",
    "vim",
    "vue",
    "yaml",
    "zig",
}

function module.apply()
    ts.setup {
        ensure_installed = ensure_installed,
        sync_install = true,
        highlight = {enable = true, additional_vim_regex_highlighting = false},
        incremental_selection = {
            enable = true,
            keymaps = {
                init_selection = module.selection_init,
                node_incremental = module.selection_inc,
                scope_incremental = module.selection_inc_scope,
                node_decremental = module.selection_dec,
            },
        },
        indent = {enable = false},
        playground = {
            enable = true,
            disable = {},
            updatetime = 25,         -- Debounced time for highlighting nodes in the playground from source code
            persist_queries = false, -- Whether the query persists across vim sessions
            keybindings = {
                toggle_query_editor = "o",
                toggle_hl_groups = "i",
                toggle_injected_languages = "t",
                toggle_anonymous_nodes = "a",
                toggle_language_display = "I",
                focus_language = "f",
                unfocus_language = "F",
                update = "R",
                goto_node = "<cr>",
                show_help = "?",
            },
        },
    }
    require("Comment").setup(
        {
            toggler = {line = module.comment_toggle_line, block = module.comment_toggle_block},
            opleader = {line = module.comment_line, block = module.comment_block},
            mappings = {basic = true, extra = false, extended = false},
        }
    )
    require("treesitter-context").setup {
        enable = true,
        throttle = true,
        max_lines = 0,
        patterns = {
            -- For all filetypes
            -- Note that setting an entry here replaces all other patterns for this entry.
            -- By setting the 'default' entry below, you can control which nodes you want to
            -- appear in the context window.
            default = {"class", "function", "method", "for", "while", "if", "switch", "case"},
            -- Example for a specific filetype.
            -- If a pattern is missing, *open a PR* so everyone can benefit.
            --   rust = {
            --       'impl_item',
            --   },
            json = {"object", "array"},
            yaml = {"block_mapping_pair"},
        },
        exact_patterns = {
            -- Example for a specific filetype with Lua patterns
            -- Treat patterns.rust as a Lua pattern (i.e "^impl_item$" will
            -- exactly match "impl_item" only)
            -- rust = true,
        },
    }
end

return module
