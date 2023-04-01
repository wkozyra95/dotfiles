local ls = require "luasnip"

local i = ls.insert_node
local fmt = require("luasnip.extras.fmt").fmt

return {req = fmt("import {} from \"{}\"", {i(1), i(2)})}
