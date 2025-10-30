return {
  {
    "williamboman/mason.nvim",
    config = function()
      require("mason").setup()
    end,
  },
  {
    "williamboman/mason-lspconfig.nvim",
    config = function()
      require("mason-lspconfig").setup({
        ensure_installed = {
          "lua_ls",
          "gopls",
          "yamlls",
          "jsonls",
          "sqlls",
          "clangd",
          "jdtls",
          -- JS/TS LSP (new):
          "ts_ls",    -- new name in recent lspconfig
          -- "tsserver", -- uncomment if you're on older lspconfig
        },
      })
    end,
  },
  {
    "neovim/nvim-lspconfig",
    dependencies = { "saghen/blink.cmp" },
    config = function()
      local capabilities = require("blink.cmp").get_lsp_capabilities()
      local lspconfig = require("lspconfig")
      local util = lspconfig.util

      vim.keymap.set("n", "<leader>f", function() vim.lsp.buf.format() end)
      vim.keymap.set("n", "gd", vim.lsp.buf.definition, {})
      vim.keymap.set("n", "K", vim.lsp.buf.hover, {})
      vim.keymap.set("n", "<leader>ca", vim.lsp.buf.code_action, {})

      -- Lua
      lspconfig.lua_ls.setup({
        capabilities = capabilities,
        settings = {
          Lua = {
            diagnostics = { globals = { "vim" } },
            workspace = { library = vim.api.nvim_get_runtime_file("", true) },
            telemetry = { enable = false },
          },
        },
      })

      -- Go
      lspconfig.gopls.setup({ capabilities = capabilities })

      -- Java
      lspconfig.jdtls.setup({ capabilities = capabilities })

      -- YAML
      lspconfig.yamlls.setup({ capabilities = capabilities })

      -- JSON
      lspconfig.jsonls.setup({ capabilities = capabilities })

      -- SQL
      lspconfig.sqlls.setup({ capabilities = capabilities })

      -- C/C++
      lspconfig.clangd.setup({
        cmd = {
          "clangd",
          "--background-index",
          "--pch-storage=memory",
          "--all-scopes-completion",
          "--pretty",
          "--header-insertion=never",
          "-j=4",
          "--inlay-hints",
          "--header-insertion-decorators",
          "--function-arg-placeholders",
          "--completion-style=detailed",
        },
        filetypes = { "c", "cpp", "objc", "objcpp" },
        root_dir = lspconfig.util.root_pattern("src"),
        init_option = { fallbackFlags = { "-std=c++2a" } },
        capabilities = capabilities,
        single_file_support = true,
      })

      -- ✨ JavaScript/TypeScript (new)
      local tsserver = lspconfig.ts_ls or lspconfig.tsserver
      if tsserver then
        tsserver.setup({
          capabilities = capabilities,
          filetypes = {
            "javascript", "javascriptreact", "javascript.jsx",
            "typescript", "typescriptreact", "typescript.tsx",
          },
          root_dir = util.root_pattern("package.json", "tsconfig.json", "jsconfig.json", ".git"),
          single_file_support = true,
          settings = {
            javascript = { inlayHints = { includeInlayParameterNameHints = "all" } },
            typescript = { inlayHints = { includeInlayParameterNameHints = "all" } },
          },
        })
      end
    end,
  },
}

