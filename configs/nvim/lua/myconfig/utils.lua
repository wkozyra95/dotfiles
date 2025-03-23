local module = {}

local Job = require("plenary.job")
local Path = require("plenary.path")
local base64 = require("myconfig.base64")

local function do_apply_keymap(schema, prefix, default_options)
    if schema[1] and (type(schema[1]) == "string" or type(schema[1]) == "function") then
        vim.keymap.set(
            schema[2] or "", prefix, schema[1],
            vim.tbl_extend("force", default_options, schema[3] or {})
        )
    elseif schema.__is_ref then
        schema.module[schema.field] = prefix
    else
        for key, subschema in pairs(schema) do
            if type(key) == "number" then
                do_apply_keymap(subschema, prefix, default_options)
            else
                do_apply_keymap(subschema, prefix .. key, default_options)
            end
        end
    end
end

function module.ref(mod, field) return {module = mod, field = field, __is_ref = true} end

function module.apply_keymap(schema, default_options) do_apply_keymap(schema, "", default_options) end

function module.run(input)
    local args = {}
    for _, v in ipairs(input) do
        table.insert(args, v)
    end

    local command = table.remove(args, 1)

    ---@diagnostic disable-next-line: missing-fields
    Job:new {
        command = command,
        args = args,

        cwd = input.cwd,

        on_stdout = vim.schedule_wrap(function(_, data) print(command, ":", data) end),
        on_stderr = vim.schedule_wrap(function(_, data) print(command, "[err]:", data) end),
        skip_validation = input.skip_validation or true,
    }:sync(60000)
end

function module.run_async(input)
    local args = {}
    for _, v in ipairs(input) do
        table.insert(args, v)
    end

    local command = table.remove(args, 1)

    ---@diagnostic disable-next-line: missing-fields
    Job:new {
        command = command,
        args = args,

        cwd = input.cwd,

        on_stdout = vim.schedule_wrap(function(_, data) print(command, ":", data) end),
        on_stderr = vim.schedule_wrap(function(_, data) print(command, "[err]:", data) end),
        skip_validation = input.skip_validation or true,
    }:start()
end

function module.find_recursive(filename, search_start)
    local parent_paths = Path:new(search_start):parents()
    for _, path in pairs(parent_paths) do
        if Path:new(path, filename):exists() then
            return Path:new(path, filename)
        end
    end
end

function module.once(func)
    local already_called = false
    local cached_result = nil
    return function()
        if already_called then
            return cached_result
        end
        cached_result = func()
        already_called = true
        return cached_result
    end
end

function module.get_dir_from_path(path)
    if 1 == vim.fn.isdirectory(path) then
        return path
    else
        return string.gsub(path, "/[^/]+$", "")
    end
end

function module.get_buf_dir() return module.get_current_buf_dir(0) end

function module.get_current_buf_dir(bufnr)
    local current_buffer_path = vim.api.nvim_buf_get_name(bufnr)
    if current_buffer_path == "" then
        vim.notify("buffer does not have any location")
        return vim.fn.getcwd()
    end
    return module.get_dir_from_path(current_buffer_path)
end

function module.rpc_run(action)
    local json_string = vim.json.encode(action)
    local encoded = base64.encode(json_string)
    ---@diagnostic disable-next-line: missing-fields
    local job = Job:new {
        command = "mycli",
        args = {"api", encoded},

        on_stderr = vim.schedule_wrap(function(_, data)
            if data then
                print("[err]:", data)
            end
        end),
        skip_validation = true,
    }
    job:sync(60000)
    local stdout = job:result()
    if job.code ~= 0 then
        vim.notify("action " .. action.name .. " failed with exit code " .. job.code)
        return {is_error = true, stderr = job:stderr_result()}
    end
    return vim.json.decode(stdout[#stdout])
end

function module.rpc_start(action, cb)
    local json_string = vim.json.encode(action)
    local encoded = base64.encode(json_string)
    ---@diagnostic disable-next-line: missing-fields
    return Job:new {
        command = "mycli",
        args = {"api", encoded},

        on_stderr = vim.schedule_wrap(function(_, data)
            if data then
                print("[err]:", data)
            end
        end),
        on_exit = vim.schedule_wrap(
            function(self, return_val)
                local stdout = self:result()
                if return_val ~= 0 then
                    vim.notify("action " .. action.name .. " failed with exit code " .. return_val)
                    cb({}, {is_error = true, stderr = self:stderr_result()})
                end
                cb(vim.json.decode(stdout[1]))
            end
        ),
        skip_validation = true,
    }:start()
end

function module.max(a, b)
    if a > b then
        return a
    else
        return b
    end
end

return module
