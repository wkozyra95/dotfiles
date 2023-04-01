local module = {}

local finders = require("telescope.finders")
local pickers = require("telescope.pickers")
local actions = require "telescope.actions"
local action_state = require "telescope.actions.state"
local conf = require("telescope.config").values
local tel = require("telescope")
local _ = require("myconfig.utils")

local function entry_maker(entry)
    return {value = entry, path = entry.path, display = entry.name, ordinal = entry.name}
end

function module.set_workspace(path)
    vim.cmd.cd(path)
    vim.cmd.wall()
    local buffers = vim.api.nvim_list_bufs()
    for _, b in ipairs(buffers) do
        local bufname = vim.api.nvim_buf_get_name(b)
        if bufname ~= "" and string.sub(bufname, 0, 8) ~= "noice://" then
            vim.api.nvim_buf_delete(b, {})
        end
    end
    tel.extensions.file_browser.file_browser({hidden = true, respect_gitignore = true})
    vim.api.nvim_feedkeys(vim.api.nvim_replace_termcodes("<esc>", true, false, true), "n", false)
end

module.list = {}
module.current = {}

local function create_workspaces_list()
    return function()
        pickers.new(
            {}, {
                prompt_title = "Switch workspace",
                finder = finders.new_table {results = module.list, entry_maker = entry_maker},
                sorter = conf.generic_sorter({}),
                previewer = conf.file_previewer({}),
                attach_mappings = function(prompt_buf)
                    actions.select_default:replace(
                        function()
                            actions.close(prompt_buf)
                            local entry = action_state.get_selected_entry()
                            module.set_workspace(entry.value.path)
                        end
                    )
                    return true;
                end,
            }
        ):find()
    end
end

module.switch_workspace = create_workspaces_list()

module.switch_to_current_dir = function() module.set_workspace(_.get_current_buf_dir(0)) end

module.open_terminal_root = function()
    _.run_async({"alacritty", "--working-directory", vim.fn.getcwd()})
end

module.open_terminal_current = function()
    local current_buf_dir = _.get_current_buf_dir(0)
    _.run_async({"alacritty", "--working-directory", current_buf_dir})
end

local function get_current_workspace(workspaces)
    local cwd = vim.fn.getcwd()
    local result
    for _, w in ipairs(workspaces) do
        if #(w.path) <= #cwd and string.sub(cwd, 0, #(w.path)) == w.path then
            if not result or #(result.path) < #(w.path) then
                result = w
            end
        end
    end
    return result or {vim = {}}
end

local function show_workspace_info()
    if (module.current.name) then
        vim.notify(vim.inspect(module.current))
    else
        vim.notify("No workspace found at " .. vim.fn.getcwd())
    end
end

module.apply = function(fn)
    vim.api.nvim_create_user_command("WorkspaceInfo", show_workspace_info, {nargs = 0})
    _.rpc_start(
        {name = "workspaces:list"}, vim.schedule_wrap(
            function(w)
                module.list = w
                module.current = get_current_workspace(w)
                fn()
            end
        )
    )
end

return module
