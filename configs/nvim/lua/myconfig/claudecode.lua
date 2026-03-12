local module = {}

local cc = require("claudecode")

function module.apply()
    cc.setup({
        -- Top-level aliases are supported and forwarded to terminal config
        -- git_repo_cwd = true,
        terminal_cmd = "~/.local/bin/claude"
    })
end

return module;
