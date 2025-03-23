local module = {
    settings = {
        ["rust-analyzer"] = {
            cargo = {
                features = "all",
                target = nil
            }
        }
    }
}

function module.rust_analyzer_config()
    return {settings = module.settings}
end

local function update_config()
    local client = vim.lsp.get_clients({name = "rust_analyzer"})[1]
    if not client then
        return
    end

    client.notify("workspace/didChangeConfiguration", {settings = module.settings})
end

function module.actions()
    if vim.bo.filetype ~= "rust" then
        return {}
    end

    if not module.settings["rust-analyzer"].cargo.target then
        return {
            rust_use_wasm_target = {
                name = "[rust] use WASM target_arch",
                fn = function()
                    module.settings["rust-analyzer"].cargo.target = "wasm32-unknown-unknown"
                    module.settings["rust-analyzer"].cargo.features = nil
                    module.settings["rust-analyzer"].cargo.noDefaultFeatures = true
                    update_config()
                end
            }
        }
    else
        return {
            rust_use_default_target = {
                name = "[rust] use default target_arch",
                fn = function()
                    module.settings["rust-analyzer"].cargo.target = nil
                    module.settings["rust-analyzer"].cargo.features = "all"
                    module.settings["rust-analyzer"].cargo.noDefaultFeatures = false
                    update_config()
                end
            }
        }
    end
end

return module;
