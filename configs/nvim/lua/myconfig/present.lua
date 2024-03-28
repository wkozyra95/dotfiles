local module = {}

local present_context = {
    presenting = false,
    hl = {}
}

local function present_option_toggle(option, value)
    if present_context.presenting then
        vim.opt[option] = present_context[option]
    else
        present_context[option] = vim.opt[option]
        vim.opt[option] = value
    end
end

local function hl_toggle(name, options)
    if present_context.presenting then
        vim.api.nvim_set_hl(0, name, present_context.hl[name])
    else
        local old_hl = vim.api.nvim_get_hl(0, {name = name})
        vim.api.nvim_set_hl(0, name, options)
        present_context.hl[name] = old_hl
    end
end

function module.present_toggle()
    present_option_toggle("showtabline", 0)
    present_option_toggle("laststatus", 0)
    present_option_toggle("number", false)
    present_option_toggle("relativenumber", false)
    present_option_toggle("signcolumn", "yes:2")
    present_option_toggle("ruler", false)
    present_option_toggle("cursorline", false)
    present_option_toggle("winbar", "")
    present_option_toggle("fillchars", {eob = " "})
    present_option_toggle("spell", false)

    hl_toggle("Transparent", {bg = "#000000", fg = "#000000", blend = 100})
    hl_toggle("SignColumn", {bg = "NONE", fg = "NONE", blend = 100})

    vim.cmd [[GitGutterToggle]]
    if present_context.presenting then
        vim.cmd [[NoiceEnable]]
    else
        vim.cmd [[NoiceDisable]]
    end

    present_option_toggle("guicursor",
        "n-v-c-sm:block-Transparent,i-ci-ve:ver25-Transparent,r-cr-o:hor20-Transparent")

    present_context.presenting = not present_context.presenting
    module.on_buf_enter_hook();
end

function module.is_presenting()
    return present_context.presenting
end

function module.on_buf_enter_hook()
    if present_context.presenting then
        vim.diagnostic.disable()
    else
        vim.diagnostic.enable()
    end
end

return module
