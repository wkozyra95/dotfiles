local module = {}
local action = require("telescope.actions")
local action_state = require("telescope.actions.state")
local tel = require("telescope")
local finders = require "telescope.finders"
local pickers = require("telescope.pickers")
local conf = require("telescope.config").values
local make_entry = require("telescope.make_entry")
local workspaces = require("myconfig.workspaces")
local _ = require("myconfig.utils")

module.state = {
    show_ignored = false
}

function module.file_browser_root()
    tel.extensions.file_browser.file_browser(
        {
            hidden = true,
            respect_gitignore = not module.state.show_ignored,
            git_status = false
        }
    )
    vim.api.nvim_feedkeys(vim.api.nvim_replace_termcodes("<esc>", true, false, true), "n", false)
end

function module.file_browser_current_dir()
    tel.extensions.file_browser.file_browser(
        {
            cwd = _.get_current_buf_dir(0),
            hidden = true,
            respect_gitignore = not module.state.show_ignored,
            git_status = false,
        }
    )
    vim.api.nvim_feedkeys(vim.api.nvim_replace_termcodes("<esc>", true, false, true), "n", false)
end

local function swich_directory_action(bufnr)
    local path = action_state.get_selected_entry().cwd
    action.close(bufnr)
    workspaces.set_workspace(path)
end

function module.live_grep(opts)
    opts = opts or {}
    opts.cwd = opts.cwd and vim.fn.expand(opts.cwd) or vim.loop.cwd()
    local no_ignore = opts.no_ignore or false
    local hidden = opts.hidden or false

    local custom_grep = finders.new_async_job {
        command_generator = function(prompt)
            if not prompt or prompt == "" then
                return nil
            end

            local prompt_split = vim.split(prompt, "  ")

            local args = {"rg"}
            if prompt_split[1] then
                table.insert(args, "-e")
                table.insert(args, prompt_split[1])
            end

            if prompt_split[2] then
                table.insert(args, "-g")
                table.insert(args, "**/*" .. prompt_split[2] .. "{*,*/**}")
            end

            return vim.tbl_flatten {
                args,
                {
                    "--color=never",
                    "--no-heading",
                    "--with-filename",
                    "--line-number",
                    "--column",
                    "--smart-case",
                },
                no_ignore and {"--no-ignore"} or {},
                hidden and {"--hidden"} or {},
            }
        end,
        entry_maker = make_entry.gen_from_vimgrep(opts),
        cwd = opts.cwd,
    }

    pickers.new(
        opts, {
            debounce = 100,
            prompt_title = "Live Grep",
            finder = custom_grep,
            previewer = conf.grep_previewer(opts),
            sorter = require("telescope.sorters").empty(),
            attach_mappings = function(_, map)
                local toggle_ignore = function() no_ignore = not no_ignore end
                map("n", "t", toggle_ignore)
                map("i", "<c-t>", toggle_ignore)
                map("n", "<c-t>", toggle_ignore)
                return true
            end,
        }
    ):find()
end

function module.apply()
    local fb_actions = tel.extensions.file_browser.actions
    require("telescope._extensions.file_browser.config").values.mappings = {
        ["i"] = {},
        ["n"] = {},
    }
    tel.setup {
        defaults = {
            mappings = {
                n = {
                    ["<C-j>"] = action.move_selection_next,
                    ["<C-k>"] = action.move_selection_previous,
                },
                i = {
                    ["<C-j>"] = action.move_selection_next,
                    ["<C-k>"] = action.move_selection_previous,
                },
            },
            layout_config = {prompt_position = "top"},
            sorting_strategy = "ascending",
        },
        extensions = {
            fzf = {
                fuzzy = true,
                override_generic_sorter = true,
                override_file_sorter = true,
                case_mode = "smart_case",
            },
            file_browser = {
                hijack_netrw = true,
                mappings = {
                    ["i"] = {
                        ["<C-e>"] = fb_actions.goto_cwd,
                        ["<C-p>"] = require("myconfig.actions").actions.find_files.fn,

                        ["<left>"] = fb_actions.goto_parent_dir,
                        ["<C-h>"] = fb_actions.goto_parent_dir,
                        ["<right>"] = action.select_default,
                        ["<C-l>"] = action.select_default,
                        ["<Esc><Esc>"] = action.close,
                    },
                    ["n"] = {
                        ["<C-e>"] = fb_actions.goto_cwd,
                        ["<C-p>"] = require("myconfig.actions").actions.find_files.fn,
                        ["<space>ff"] = require("myconfig.actions").actions.grep_files.fn,

                        ["c"] = fb_actions.create,
                        ["r"] = fb_actions.rename,
                        ["m"] = fb_actions.move,
                        ["y"] = fb_actions.copy,
                        ["d"] = fb_actions.remove,
                        ["e"] = fb_actions.goto_cwd,

                        ["<left>"] = fb_actions.goto_parent_dir,
                        ["<C-h>"] = fb_actions.goto_parent_dir,
                        ["h"] = fb_actions.goto_parent_dir,
                        ["<right>"] = action.select_default,
                        ["<C-l>"] = action.select_default,
                        ["l"] = action.select_default,
                        ["<space>sd"] = swich_directory_action,

                        ["<Esc>"] = false,
                        ["<Esc><Esc>"] = action.close,
                    },
                },
            },
        },
    }
    if not vim.fn.has("macunix") then
        tel.load_extension("fzf")
    end
    tel.load_extension("file_browser")
end

return module
