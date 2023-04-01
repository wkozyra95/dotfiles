local module = {}

function module.apply()
    require("noice").setup(
        {
            views = {
                cmdline_popup = {
                    position = {row = "50%", col = "50%"},
                    size = {width = 120, height = "auto"},
                },
                popupmenu = {
                    relative = "editor",
                    position = {row = "60%", col = "50%"},
                    size = {width = 60, height = 10},
                    border = {style = "rounded", padding = {0, 1}},
                    win_options = {
                        winhighlight = {Normal = "Normal", FloatBorder = "DiagnosticInfo"},
                    },
                },
            },
            lsp = {progress = {enabled = false}},
            messages = {enabled = false},
        }
    )
end

return module
