local workspace = require("myconfig.workspaces")

local module = {}

function module.yamlls_config()
    local schemas = {
        ["https://json.schemastore.org/kustomization.json"] = {
            "kustomization.yaml",
            "kustomization.yml"
        },
        ["https://json.schemastore.org/github-action.json"] = {
            ".github/actions/*.yml",
        },
        ["https://json.schemastore.org/github-workflow.json"] = {
            ".github/workflows/*.yml",
        },
        ["https://json.schemastore.org/circleciconfig.json"] = {
            ".circleci/config.yml",
        },
        kubernetes = "k8s/**/*.yaml",
    }
    if workspace.current and workspace.current.vim.yml_schemas then
        schemas = vim.list_extend(schemas, workspace.current.vim.yml_schemas)
    end
    return {
        settings = {
            yaml = {
                schemas = schemas,
                format = {
                    enable = true,
                },
                hover = true,
                schemaDownload = {enable = true},
                validate = true,
                completion = true,
            },
        },
    }
end

return module
