-- custom/init.lua
local opt = vim.opt

-- Add any custom options here
opt.relativenumber = true

-- custom/configs/lspconfig.lua
local on_attach = require("plugins.configs.lspconfig").on_attach
local capabilities = require("plugins.configs.lspconfig").capabilities

local lspconfig = require("lspconfig")

lspconfig.gopls.setup{
  on_attach = on_attach,
  capabilities = capabilities,
  settings = {
    gopls = {
      analyses = {
        unusedparams = true,
      },
      staticcheck = true,
    },
  },
}

-- custom/configs/null-ls.lua
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

-- custom/plugins.lua
local plugins = {
  {
    "neovim/nvim-lspconfig",
    config = function()
      require "plugins.configs.lspconfig"
      require "custom.configs.lspconfig"
    end,
  },
  {
    "jose-elias-alvarez/null-ls.nvim",
    ft = "go",
    opts = function()
      return require "custom.configs.null-ls"
    end,
  },
  {
    "olexsmir/gopher.nvim",
    ft = "go",
    config = function(_, opts)
      require("gopher").setup(opts)
    end,
    build = function()
      vim.cmd [[silent! GoInstallDeps]]
    end,
  },
  {
    "mfussenegger/nvim-dap",
    init = function()
      require("core.utils").load_mappings("dap")
    end
  },
  {
    "leoluz/nvim-dap-go",
    ft = "go",
    dependencies = "mfussenegger/nvim-dap",
    config = function(_, opts)
      require("dap-go").setup(opts)
    end
  }
}

return plugins

-- custom/mappings.lua
local M = {}

M.gopher = {
  plugin = true,
  n = {
    ["<leader>gsj"] = { "<cmd>GoTagAdd json<cr>", "Add json struct tags" },
    ["<leader>gsy"] = { "<cmd>GoTagAdd yaml<cr>", "Add yaml struct tags" },
  },
}

M.dap = {
  plugin = true,
  n = {
    ["<leader>db"] = { "<cmd>DapToggleBreakpoint<cr>", "Add breakpoint" },
    ["<leader>dus"] = {
      function ()
        local widgets = require('dap.ui.widgets');
        local sidebar = widgets.sidebar(widgets.scopes);
        sidebar.open();
      end,
      "Open debugging sidebar"
    }
  }
}

return M
