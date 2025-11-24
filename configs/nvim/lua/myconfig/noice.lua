local module = {}

function module.apply()
    require("noice").setup(
        {
            views = {
                cmdline_popup = {
                    position = {row = 30, col = "50%"},
                    size = {width = "auto", height = "auto"},
                },
                popupmenu = {
                    relative = "editor",
                    position = {row = 35, col = "50%"},
                    size = {width = 60, height = 10},
                    border = {style = "rounded", padding = {0, 1}},
                    win_options = {
                        winhighlight = {Normal = "Normal", FloatBorder = "DiagnosticInfo"},
                    },
                },
                mini = {
                    backend = "mini",
                    relative = "editor",
                    align = "message-right",
                    timeout = 2000,
                    reverse = true,
                    focusable = false,
                    position = {
                        row = -3,
                        col = "100%",
                        -- col = 0,
                    },
                    size = {
                        width = "auto",
                        height = "auto",
                        max_height = 10,
                    },
                    border = {
                        style = "none",
                    },
                    zindex = 60,
                    win_options = {
                        winbar = "",
                        foldenable = false,
                        winblend = 90,
                        winhighlight = {
                            Normal = "NoiceMini",
                            IncSearch = "",
                            CurSearch = "",
                            Search = "",
                        },
                    },
                },
            },
            presets = {
                bottom_search = false,         -- use a classic bottom cmdline for search
                command_palette = false,       -- position the cmdline and popupmenu together
                long_message_to_split = false, -- long messages will be sent to a split
                inc_rename = false,            -- enables an input dialog for inc-rename.nvim
                lsp_doc_border = false,        -- add a border to hover docs and signature help
            },
            messages = {
                enabled = true,            -- enables the Noice messages UI
                view = "notify",           -- default view for messages
                view_error = "notify",     -- view for errors
                view_warn = "notify",      -- view for warnings
                view_history = "messages", -- view for :messages
                view_search = false,       -- view for search count messages. Set to `false` to disable
            },
            routes = {
                -- use :Noice debug
                -- kind See :h ui-messages

                -- route to bottom-right corner
                {filter = {event = "notify", find = "method textDocument/hover is not supported"},   view = "mini"},
                {filter = {event = "notify", kind = "info"},                                         view = "mini"}, -- e.g. LSP info messages
                {filter = {event = "msg_show", kind = "emsg", find = "E486"},                        view = "mini"}, -- pattern not found

                -- skip
                {filter = {event = "msg_show", find = "is deprecated"},                              skip = true}, -- vim deprecated api
                {filter = {event = "msg_show", kind = "", find = "lines yanked"},                    skip = true},
                {filter = {event = "msg_show", kind = "", find = "more lines"},                      skip = true},
                {filter = {event = "msg_show", kind = "", find = "fewer lines"},                     skip = true},
                {filter = {event = "msg_show", kind = "echomsg", find = "No more valid diagnostic"}, skip = true},
                --{filter = {find = "No signature help"},                      skip = true},
                --{filter = {find = "E37"},                                    skip = true},
            },
            lsp = {
                progress = {
                    enabled = true,
                    -- Lsp Progress is formatted using the builtins for lsp_progress. See config.format.builtin
                    -- See the section on formatting for more details on how to customize.
                    --- @type NoiceFormat|string
                    format = "lsp_progress",
                    --- @type NoiceFormat|string
                    format_done = "lsp_progress_done",
                    throttle = 1000 / 30, -- frequency to update lsp progress message
                    view = "mini",
                },
                override = {
                    -- override the default lsp markdown formatter with Noice
                    ["vim.lsp.util.convert_input_to_markdown_lines"] = true,
                    -- override the lsp markdown formatter with Noice
                    ["vim.lsp.util.stylize_markdown"] = true,
                    -- override cmp documentation with Noice (needs the other options to work)
                    ["cmp.entry.get_documentation"] = true,
                },
                hover = {
                    enabled = true,
                    silent = false, -- set to true to not show a message if hover is not available
                    view = nil,     -- when nil, use defaults from documentation
                    ---@type NoiceViewOptions
                    opts = {},      -- merged with defaults from documentation
                },
                signature = {
                    enabled = true,
                    auto_open = {
                        enabled = true,
                        trigger = true, -- Automatically show signature help when typing a trigger character from the LSP
                        luasnip = true, -- Will open signature help when jumping to Luasnip insert nodes
                        throttle = 50,  -- Debounce lsp signature help request by 50ms
                    },
                    view = nil,         -- when nil, use defaults from documentation
                    ---@type NoiceViewOptions
                    opts = {},          -- merged with defaults from documentation
                },
            }

        }
    )
    require("telescope").load_extension("noice")
end

return module
