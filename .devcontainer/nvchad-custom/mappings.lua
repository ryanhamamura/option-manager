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
