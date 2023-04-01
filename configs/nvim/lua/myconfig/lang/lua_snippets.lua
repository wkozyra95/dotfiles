local ls = require "luasnip"

local i = ls.insert_node
local rep = require("luasnip.extras").rep
local fmt = require("luasnip.extras.fmt").fmt

return {
    req = fmt("local {} = require(\"{}\")", {i(1), rep(1)}),
    reqq = fmt("local {} = require(\"{}\")", {i(1), i(2)}),
}
