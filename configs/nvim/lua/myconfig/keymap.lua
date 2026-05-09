local debug = require("myconfig.debug")
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

-- # Modes
-- n  normal
-- x  visual
-- s  select
-- v  visual + select
-- o  operator-pending
-- i  insert
-- c  command-line
-- t  terminal
-- "" defaults to n + v + o

-- noremap: ignore other mappings on the rhs (use the original Vim behavior)
-- silent:  do not echo the rhs in the command line while it runs
local defaults = {noremap = true, silent = true}

local function map(lhs, rhs, mode, extra)
    local opts = extra and vim.tbl_extend("force", defaults, extra) or defaults
    vim.keymap.set(mode or "", lhs, rhs, opts)
end

-- ============================================================================
-- Single-key bindings (top-level)
-- ============================================================================
map("?", tel_builtin.current_buffer_fuzzy_find)
map("H", "gT")
map("L", "gt")
map("Y", "y$", "n")
map("Q", "<nop>", "", {noremap = false})
map("<bs>", "<C-^>", "n")
map(",,", actions.select_action())

-- ============================================================================
-- g* — LSP navigation
-- ============================================================================
map("gd", lsp.goToDefinition)
map("gD", lsp.goToDeclaration)
map("gt", lsp.goToTypeDefinition)
map("gr", lsp.references)
map("gi", lsp.goToPrev)
map("gu", lsp.goToNext)

-- ============================================================================
-- d* — delete-prefix
-- ============================================================================
map("ds", surround.remove, "n")

-- ============================================================================
-- c* — change-prefix
-- ============================================================================
map("cs", surround.replace, "n")

-- ============================================================================
-- t* — terminal/tabs
-- ============================================================================
map("tt", "<cmd>terminal<cr>", "n")

-- ============================================================================
-- <C-*> — chord prefix
-- ============================================================================
map("<C-e>", tel.file_browser_root)
map("<C-h>", lsp.onHover)
map("<C-j>", snippets.jump_back, {"i", "s"})
map("<C-k>", snippets.expand_or_jump, {"i", "s"})
map("<C-n>", "<cmd>tab split<cr>")
map("<C-p>", actions.actions.find_files.fn)
map("<C-s>", tree.start_selection, "n")
map("<C-s>", tree.expand_node, "x")
map("<C-x>", tree.shrink_node, "x")

-- ============================================================================
-- <space>* — leader
-- ============================================================================
map("<space><space>", tel.file_browser_current_dir)
map("<space><bs>", ":<Up>", "", {silent = false})

-- <space>a* — LSP actions
map("<space>aa", lsp.codeAction)
map("<space>ar", lsp.rename)
map("<space>aq", lsp.autoFix)
map("<space>af", lsp.format, "n")
map("<space>af", lsp.formatSelected, "v")
map("<space>as", actions.actions.http_request.fn)

-- <space>c* — diagnostics
map("<space>cd", lsp.showLineDiagnostics)

-- <space>d* — debug / dev
map("<space>dl", lsp.restart)
map("<space>dp", debug.playground, "", {silent = false})
map("<space>dr", debug.reload)
map("<space>ds", snippets.reload)

-- <space>f* — find
map("<space>ff", actions.actions.grep_files.fn)
map("<space>fgc", tel_builtin.git_commits)
map("<space>fgb", tel_builtin.git_branches)
map("<space>fv", actions.actions.dotfiles_search.fn)
map("<space>fn", actions.actions.notes_search.fn)
map("<space>fh", tel_builtin.help_tags)
map("<space>fb", tel_builtin.buffers)
map("<space>fr", tel_builtin.resume)

-- <space>g* — git
map("<space>gs", function() vim.cmd.vertical("G") end)
map("<space>gb", function() vim.cmd.vertical("G blame") end)
map("<space>gg", neogit.open)
map("<space>gd", git.toggle_diffview)
map("<space>gh", git.toggle_diffhistory)

-- <space>s* — workspace
map("<space>ss", workspaces.switch_workspace)
map("<space>sd", workspaces.switch_to_current_dir)

-- ============================================================================
-- Surround (visual wrap + insert auto-pair)
-- ============================================================================
map("\"", surround.surround_selection({"\""}, {"\""}), "x")
map("'", surround.surround_selection({"'"}, {"'"}), "x")
map("`", surround.surround_selection({"`"}, {"`"}), "x")
map("`", surround.auto_pair({"`"}, {"`"}), "i")
map("{", surround.surround_selection({"{"}, {"}"}), "x")
map("{", surround.auto_pair({"{"}, {"}"}), "i")
map("(", surround.surround_selection({"("}, {")"}), "x")
map("(", surround.auto_pair({"("}, {")"}), "i")
map("[", surround.surround_selection({"["}, {"]"}), "x")
map("[", surround.auto_pair({"["}, {"]"}), "i")
map("<", surround.surround_selection({"<"}, {">"}), "x")
map("}", surround.surround_selection({"{", ""}, {"", "}"}), "x")
map(")", surround.surround_selection({"(", ""}, {"", ")"}), "x")
map("]", surround.surround_selection({"[", ""}, {"", "]"}), "x")

-- ============================================================================
-- Side-effect tweaks (keep behavior, change registers / undo points)
-- ============================================================================
-- insert-mode history breakpoints (<c-g>u)
map("<space>", "<space><c-g>u", "i")
map(",", ",<c-g>u", "i")
map(".", ".<c-g>u", "i")
-- send delete to black hole register
map("c", "\"_c")
map("C", "\"_C")
map("x", "\"_x")
map("X", "\"_X")
-- preserve copy register on visual paste
map("p", "\"_dP", "v")
