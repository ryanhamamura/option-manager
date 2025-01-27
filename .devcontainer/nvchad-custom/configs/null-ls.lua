local null_ls = require("null-ls")
local b = null_ls.builtins

local sources = {
  b.formatting.gofmt,
  b.formatting.goimports,
  b.diagnostics.golangci_lint,
}

null_ls.setup {
  debug = true,
  sources = sources,
}
