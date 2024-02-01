local o = vim.opt

vim.cmd.colorscheme("gruvbox")

o.exrc = true         -- read rc files in parent directories
o.secure = true       -- exrc applies only if user is an owner

o.autowriteall = true -- autosave
o.undofile = true     -- preserve undo history between sessions
o.undodir = vim.env.HOME .. "/.cache/nvim/undo"
o.history = 5000
o.clipboard = "unnamedplus"

o.pumblend = 17         -- transparency for popup (0..100)

o.number = true         -- show numbers
o.relativenumber = true -- show relative numbers (current line is absolute)
o.scrolloff = 8         -- start scrolling n lines before buffer end
o.showmode = false      -- do not display mode under statusline

o.incsearch = true      -- highlight partial search results
o.inccommand = "split"  -- highlight + split on %s/pater/replace
o.hlsearch = true       -- highlight search results after <cr>

o.laststatus = 3        -- (2=always,3=global) statusline behavior
o.mouse = "a"           -- (a=all modes) mouse support
o.splitright = true     -- new vert split starts on right

-- what backspace can remove in insert mode
o.backspace = {
    "indent", -- indents
    "eol",    -- eol
    "start",  -- can only remove stuff added after last switch to insert mode
}

o.foldenable = false   -- do not fold parts of the code (e.g functions)

o.background = "dark"  -- selects dark theme from current scheme
o.termguicolors = true -- use real colors, term support required

-- default tab behavior
o.tabstop = 4
o.shiftwidth = 4
o.softtabstop = 4
o.expandtab = true
o.smarttab = true -- global

o.autoindent = true
o.formatoptions = o.formatoptions
    - "o"            -- O and o, don't continue comments
    + "r"            -- But do continue when pressing enter.
    + "n"            -- Indent past the formatlistpat, not underneath it.
    + "j"            -- Auto-remove comments if possible.
    - "2"            -- use indent of the second line above

o.joinspaces = false -- disable 2 spaces after join(J)

o.completeopt = {"menu", "noselect"}

o.spelllang = "en"
o.spellfile = vim.env.HOME .. "/.dotfiles/configs/nvim/spell/common.utf-8.add" .. "," ..
    vim.env.HOME .. "/.dotfiles/configs/nvim/spell/natural.utf-8.add" .. "," ..
    vim.env.HOME .. "/.dotfiles/configs/nvim/spell/blacklistonly.utf-8.add"
o.spellcapcheck = ""
o.spelloptions = {"camel"}

o.cursorline = true -- highlight current line
local group = vim.api.nvim_create_augroup("OptionsControl", {clear = true})
vim.api.nvim_create_autocmd(
    "WinLeave", {group = group, callback = function() vim.opt_local.cursorline = false end}
)
vim.api.nvim_create_autocmd(
    "WinEnter", {group = group, callback = function() vim.opt_local.cursorline = true end}
)
