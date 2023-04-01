local ls = require "luasnip"

local i = ls.insert_node
local t = ls.text_node
local rep = require("luasnip.extras").rep
local fmt = require("luasnip.extras.fmt").fmt

return {
    main = {t {"func main() {", "\t"}, i(0), t {"", "}"}},
    ef = {i(1, {"val"}), t ", err := ", i(2, {"f"}), t "(", i(3), t ")", i(0)},
    Ef = fmt("{}, {}Err := {}({})", {i(1), rep(1), i(2), i(3)}),
}
