local module = {}

local diffview = require("diffview")
local neogit = require("neogit")
local dv_actions = require("diffview.actions")
local dv_lib = require("diffview.lib")

function module.toggle_diffview()
    local view = dv_lib.get_current_view()
    if view then
        view:close()
        dv_lib.dispose_view(view)
    else
        diffview.open()
    end
end

function module.toggle_diffhistory()
    local view = dv_lib.get_current_view()
    if view then
        view:close()
        dv_lib.dispose_view(view)
    else
        diffview.file_history()
    end
end

function module.apply()
    neogit.setup({integrations = {diffview = true}})

    diffview.setup {
        diff_binaries = false,    -- Show diffs for binaries
        enhanced_diff_hl = false, -- See ':h diffview-config-enhanced_diff_hl'
        use_icons = true,         -- Requires nvim-web-devicons
        default_args = {
            -- Default args prepended to the arg-list for the listed commands
            DiffviewOpen = {},
            DiffviewFileHistory = {},
        },
        hooks = {},                                             -- See ':h diffview-config-hooks'
        key_bindings = {
            disable_defaults = true,                            -- Disable the default key bindings
            view = {
                ["<tab>"] = dv_actions["select_next_entry"],    -- Open the diff for the next file
                ["<s-tab>"] = dv_actions["select_prev_entry"],  -- Open the diff for the previous file
                ["gf"] = dv_actions["goto_file"],               -- Open the file in a new split in previous tabpage
                ["<C-w><C-f>"] = dv_actions["goto_file_split"], -- Open the file in a new split
                ["<C-w>gf"] = dv_actions["goto_file_tab"],      -- Open the file in a new tabpage
                ["<leader>e"] = dv_actions["focus_files"],      -- Bring focus to the files panel
                ["<leader>b"] = dv_actions["toggle_files"],     -- Toggle the files panel.
            },
            file_panel = {
                ["j"] = dv_actions["next_entry"],      -- Bring the cursor to the next file entry
                ["<down>"] = dv_actions["next_entry"],
                ["k"] = dv_actions["prev_entry"],      -- Bring the cursor to the previous file entry.
                ["<up>"] = dv_actions["prev_entry"],
                ["<cr>"] = dv_actions["select_entry"], -- Open the diff for the selected entry.
                ["o"] = dv_actions["select_entry"],
                ["<2-LeftMouse>"] = dv_actions["select_entry"],
                ["-"] = dv_actions["toggle_stage_entry"], -- Stage / unstage the selected entry.
                ["S"] = dv_actions["stage_all"],          -- Stage all entries.
                ["U"] = dv_actions["unstage_all"],        -- Unstage all entries.
                ["X"] = dv_actions["restore_entry"],      -- Restore entry to the state on the left side.
                ["R"] = dv_actions["refresh_files"],      -- Update stats and entries in the file list.
                ["<tab>"] = dv_actions["select_next_entry"],
                ["<s-tab>"] = dv_actions["select_prev_entry"],
                ["gf"] = dv_actions["goto_file"],
                ["<C-w><C-f>"] = dv_actions["goto_file_split"],
                ["<C-w>gf"] = dv_actions["goto_file_tab"],
                ["i"] = dv_actions["listing_style"],       -- Toggle between 'list' and 'tree' views
                ["f"] = dv_actions["toggle_flatten_dirs"], -- Flatten empty subdirectories in tree listing style.
                ["<leader>e"] = dv_actions["focus_files"],
                ["<leader>b"] = dv_actions["toggle_files"],
            },
            file_history_panel = {
                ["g!"] = dv_actions["options"],               -- Open the option panel
                ["<C-A-d>"] = dv_actions["open_in_diffview"], -- Open the entry under the cursor in a diffview
                ["y"] = dv_actions["copy_hash"],              -- Copy the commit hash of the entry under the cursor
                ["zR"] = dv_actions["open_all_folds"],
                ["zM"] = dv_actions["close_all_folds"],
                ["j"] = dv_actions["next_entry"],
                ["<down>"] = dv_actions["next_entry"],
                ["k"] = dv_actions["prev_entry"],
                ["<up>"] = dv_actions["prev_entry"],
                ["<cr>"] = dv_actions["select_entry"],
                ["o"] = dv_actions["select_entry"],
                ["<2-LeftMouse>"] = dv_actions["select_entry"],
                ["<tab>"] = dv_actions["select_next_entry"],
                ["<s-tab>"] = dv_actions["select_prev_entry"],
                ["gf"] = dv_actions["goto_file"],
                ["<C-w><C-f>"] = dv_actions["goto_file_split"],
                ["<C-w>gf"] = dv_actions["goto_file_tab"],
                ["<leader>e"] = dv_actions["focus_files"],
                ["<leader>b"] = dv_actions["toggle_files"],
            },
            option_panel = {["<tab>"] = dv_actions["select_entry"], ["q"] = dv_actions["close"]},
        },
    }
end

return module
