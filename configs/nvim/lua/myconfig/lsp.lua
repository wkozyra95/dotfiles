local module = {}

local lspkind = require("lspkind")
local lsp_config = require("lspconfig")
local cmp = require("cmp.init")
local tel = require("telescope.builtin")
local _ = require("myconfig.utils")

local filetype_mapping = {
    lua = _.once(function() return require("myconfig.lang.lua") end),
    go = _.once(function() return require("myconfig.lang.go") end),
}

--
-- @param action_name refers to method action_name in module assigned to
-- that filetype e.g. require("myconfig.lsp.lua").format()
local function lsp(opts)
    return function()
        local lsp_fn
        -- if client does not support capability
        if opts.required_capability then
            for _, client in pairs(vim.lsp.get_clients()) do
                if client.server_capabilities[opts.required_capability] == true then
                    lsp_fn = opts.lsp_func
                    break
                end
            end
        end
        if not lsp_fn and opts.fallback then
            lsp_fn = opts.fallback
        end

        -- if custom implementation exists
        local filetype = vim.api.nvim_buf_get_option(0, "filetype")
        if filetype and filetype_mapping[filetype] then
            local mod = filetype_mapping[filetype]()
            if mod and mod[opts.action_name] then
                mod[opts.action_name](lsp_fn)
                return
            end
        end
        if (lsp_fn) then
            lsp_fn()
        end
    end
end

module.goToDefinition = tel.lsp_definitions
module.goToDeclaration = vim.lsp.buf.declaration
module.goToTypeDefinition = vim.lsp.buf.type_definition
module.goToNext = vim.diagnostic.goto_next
module.goToPrev = vim.diagnostic.goto_prev
module.references = tel.lsp_references
module.onHover = vim.lsp.buf.hover
module.showLineDiagnostics = vim.diagnostic.open_float

module.format = lsp {
    lsp_func = function() vim.lsp.buf.format({async = true}) end,
    action_name = "format",
    required_capability = "documentFormattingProvider",
    fallback = function() vim.cmd.normal("gg=G") end,
}

module.formatSelected = lsp {
    lsp_func = vim.lsp.buf.range_formatting,
    action_name = "format",
    required_capability = "documentRangeFormattingProvider",
    fallback = module.format,
}

module.rename = vim.lsp.buf.rename
module.codeAction = vim.lsp.buf.code_action
module.restart = function()
    vim.lsp.stop_client(vim.lsp.get_clients())
    vim.cmd.edit()
end
module.autoFix = function() print("autoFix not supported") end

function module.lsp_setup(name, config)
    local capabilities = config.capabilities or vim.lsp.protocol.make_client_capabilities()
    capabilities = require("cmp_nvim_lsp").default_capabilities(capabilities)
    local default_config = {
        on_attach = function(client)
            if (config.on_attach) then
                config.on_attach(client);
            end
        end,
        capabilities = capabilities,
    }
    local result_config = vim.tbl_extend("force", config, default_config)
    lsp_config[name].setup(result_config)
end

function module.apply()
    vim.lsp.handlers["textDocument/publishDiagnostics"] = vim.lsp.with(
        vim.lsp.diagnostic.on_publish_diagnostics, {
            -- Enable underline, use default values
            underline = true,
            -- Enable virtual text, override spacing to 4
            virtual_text = {spacing = 4},
            -- Use a function to dynamically turn signs off
            -- and on, using buffer local variables
            -- signs = function(bufnr, client_id) return vim.bo[bufnr].show_signs == false end,
            -- Disable a feature
            update_in_insert = false,
        }
    )

    local efm_config = {
        on_attach = function(client)
            client.server_capabilities.documentFormattingProvider = true;
            client.server_capabilities.codeActionProvider = true;
        end,
        settings = {languages = {}},
        filetypes = {},
        root_dir_patterns = {".git"},
    }

    local go = require("myconfig.lang.go")
    module.lsp_setup("gopls", go.gopls_config())
    module.lsp_setup("golangci_lint_ls", go.golangci_config())
    go.attach_efm(efm_config)

    module.lsp_setup("clangd",
        {filetypes = {"c", "cpp"}, init_options = {clangdFileStatus = true}})

    local lua = require("myconfig.lang.lua")
    module.lsp_setup("lua_ls", lua.lua_ls_config())

    local typescript = require("myconfig.lang.typescript")
    module.lsp_setup("tsserver", typescript.tsserver_config())
    module.lsp_setup("eslint", typescript.eslint_config())
    module.lsp_setup("clojure_lsp", {})
    module.lsp_setup("rust_analyzer", {})

    module.lsp_setup("yamlls", require("myconfig.lang.yaml").yamlls_config())
    module.lsp_setup("jsonls", require("myconfig.lang.json").jsonls_config())

    module.lsp_setup("elixirls", require("myconfig.lang.elixir").elixirls_config())
    module.lsp_setup("ocamllsp", {})

    local cmake = require("myconfig.lang.cmake")
    module.lsp_setup("cmake", cmake.cmake_config())
    cmake.attach_efm(efm_config)

    efm_config.root_dir = lsp_config.util.root_pattern(unpack(efm_config.root_dir_patterns or {}))
    efm_config.root_dir_patterns = nil
    module.lsp_setup("efm", efm_config)

    lspkind.init()

    ---@diagnostic disable-next-line: missing-fields
    cmp.setup {
        snippet = {expand = function(args) require("luasnip").lsp_expand(args.body) end},
        ---@diagnostic disable-next-line: missing-fields
        completion = {completeopt = "menu,noselect"},
        preselect = cmp.PreselectMode.None,
        mapping = {
            ["<C-d>"] = cmp.mapping(cmp.mapping.scroll_docs(-4), {"i", "c"}),
            ["<C-f>"] = cmp.mapping(cmp.mapping.scroll_docs(4), {"i", "c"}),
            ["<C-Space>"] = cmp.mapping.complete(),
            ["<C-e>"] = cmp.mapping.close(),
            ["<CR>"] = cmp.mapping.confirm({select = false, behavior = cmp.ConfirmBehavior.Insert}),
            ["<tab>"] = cmp.mapping.confirm({select = true, behavior = cmp.ConfirmBehavior.Insert}),
            ["<down>"] = cmp.mapping.select_next_item({behavior = cmp.SelectBehavior.Insert}),
            ["<up>"] = cmp.mapping.select_prev_item({behavior = cmp.SelectBehavior.Insert}),
        },
        sources = {
            {name = "luasnip"},
            {name = "nvim_lsp"},
            {name = "buffer",  keyword_length = 5},
            {name = "path"},
        },
        ---@diagnostic disable-next-line: missing-fields
        formatting = {
            format = lspkind.cmp_format(
                {with_text = false, menu = ({buffer = "[Buffer]", nvim_lsp = "[LSP]"})}
            ),
        },
        experimental = {ghost_text = true},
    }
end

return module
