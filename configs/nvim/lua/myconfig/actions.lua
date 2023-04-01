local finders = require("telescope.finders")
local pickers = require("telescope.pickers")
local action_state = require "telescope.actions.state"
local tel_actions = require("telescope.actions")
local tel_themes = require("telescope.themes")
local conf = require("telescope.config").values
local tel_builtin = require("telescope.builtin")
local _ = require("myconfig.utils")

local tel = require("myconfig.telescope")
local workspaces = require("myconfig.workspaces")
local lsp = require("myconfig.lsp")

local homeDir = vim.env.HOME

local actions = {
    lsp_code_actions = {name = "[lsp] code action", fn = lsp.codeAction},
    lsp_rename = {name = "[lsp] rename", fn = lsp.rename},
    shell_root = {name = "[shell] root dir", fn = workspaces.open_terminal_root},
    shell_cwd = {name = "[shell] current dir", fn = workspaces.open_terminal_current},
    workspace_switch = {name = "workspace switch", fn = workspaces.switch_workspace},
    workspace_switch_current = {
        name = "workspace switch to current dir",
        fn = workspaces.switch_to_current_dir,
    },
    format = {name = "format", fn = lsp.format},
    format_range = {name = "format range", fn = lsp.formatSelected},
    restart = {name = "restart", fn = lsp.restart},
    find_files = {
        name = "find files",
        fn = function() tel_builtin.find_files({no_ignore = tel.state.show_ignored, hidden = true}) end
    },
    grep_files = {
        name = "grep all",
        fn = function() tel.live_grep({no_ignore = tel.state.show_ignored, hidden = true}) end,
    },
    dotfiles_go = {
        name = "dotfiles",
        fn = function() workspaces.set_workspace(homeDir .. "/.dotfiles") end,
    },
    notes_search = {
        name = "search notes",
        fn = function()
            tel_builtin.find_files({cwd = "~/notes", prompt_title = "Notes", hidden = true})
        end,
    },
    dotfiles_search = {
        name = "search dotfiles",
        fn = function()
            tel_builtin.find_files({cwd = "~/.dotfiles", prompt_title = "Dot files", hidden = true})
        end,
    },
    filebrowser_noingore = {name = "file browser no ignore", fn = tel.file_browser_current_noignore},
    git_commit = {
        name = "git add -A && git commit",
        fn = function()
            vim.cmd.G("add -A")
            vim.cmd.G("commit")
        end,
    },
    git_commit_amend = {
        name = "git add -A && git commit --amend",
        fn = function()
            vim.cmd.G("add -A")
            vim.cmd.G("commit --amend")
        end,
    },
    toggle_ignored = {
        name = "toggle ignored files",
        fn = function()
            if (tel.state.show_ignored) then
                vim.notify("hiding ignored files")
                tel.state.show_ignored = false
            else
                vim.notify("showing ignored files")
                tel.state.show_ignored = true
            end
        end,
    },
    toggle_diagnostics = {
        name = "[lsp] toggle diagnostics",
        fn = (function()
            local enabled = true;
            return function()
                if (enabled) then
                    vim.diagnostic.disable()
                else
                    vim.diagnostic.enable()
                end
                enabled = not enabled
            end
        end)(),
    },
}

local function create_local_action(action)
    return {
        id = action.id,
        name = action.name,
        fn = function()
            local command = action.args
            command.cwd = action.cwd
            _.run_async(command)
        end,
    }
end

local function select_action()
    local function entry_maker(entry)
        return {value = entry, display = entry.name, ordinal = entry.id}
    end

    local actions_list = {}
    for k, v in pairs(workspaces.current.vim.actions or {}) do
        table.insert(actions_list, create_local_action(v))
    end
    for k, v in pairs(actions) do
        table.insert(actions_list, vim.tbl_extend("force", v, {id = k}))
    end
    return function()
        pickers.new(
            {}, vim.tbl_extend(
                "force", {
                    prompt_title = "Run action",
                    finder = finders.new_table({results = actions_list, entry_maker = entry_maker}),
                    sorter = conf.generic_sorter({}),
                    attach_mappings = function(prompt_buf)
                        tel_actions.select_default:replace(
                            function()
                                tel_actions.close(prompt_buf)
                                local entry = action_state.get_selected_entry()
                                entry.value.fn()
                            end
                        )
                        return true;
                    end,
                }, tel_themes.get_ivy()
            )
        ):find()
    end
end

return {actions = actions, select_action = select_action}
