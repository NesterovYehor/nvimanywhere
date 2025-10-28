-- bootstrap lazy.nvim --------------------------------------------------------
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({ "git", "clone", "--filter=blob:none", "--single-branch",
    "https://github.com/folke/lazy.nvim.git", lazypath })
end
vim.opt.rtp:prepend(lazypath)

-- core opts ------------------------------------------------------------------
vim.g.mapleader = " "
vim.opt.termguicolors = true
vim.opt.number = true
vim.opt.relativenumber = false
vim.opt.signcolumn = "yes"
vim.opt.mouse = "a"
vim.opt.updatetime = 200
vim.opt.ignorecase = true
vim.opt.smartcase = true
vim.opt.expandtab = true
vim.opt.shiftwidth = 2
vim.opt.tabstop = 2
vim.opt.completeopt = "menu,menuone,noselect"

-- plugins --------------------------------------------------------------------
require("lazy").setup({
  -- theme
  {
    "folke/tokyonight.nvim",
    lazy = false,
    priority = 1000,
    opts = { style = "night" },
    config = function() vim.cmd.colorscheme("tokyonight") end,
  },

  -- telescope
  { "nvim-lua/plenary.nvim" },
  { "nvim-telescope/telescope.nvim" },

  -- treesitter
  {
    "nvim-treesitter/nvim-treesitter",
    build = ":TSUpdate",
    opts = {
      ensure_installed = { "lua", "go", "javascript", "html", "css", "json" },
      highlight = { enable = true },
      indent = { enable = true },
    },
    config = function(_, opts)
      require("nvim-treesitter.configs").setup(opts)
    end,
  },

  -- Mason (manages LSP servers & tools)
  { "williamboman/mason.nvim", build = ":MasonUpdate", config = true },
  {
    "williamboman/mason-lspconfig.nvim",
    opts = {
      ensure_installed = {
        "gopls",
        "lua_ls",
        "tsserver",   -- (use "ts_ls" in lspconfig if available; Mason name stays tsserver)
        "html", "cssls", "jsonls",
      },
      automatic_installation = false, -- we prebake in Docker
    },
  },
  {
    "WhoIsSethDaniel/mason-tool-installer.nvim",
    opts = function()
      local prebaked = (vim.env.MASON_PREBAKED == "1")
      return {
        ensure_installed = {
          "gopls",
          "lua-language-server",
          "typescript-language-server",
          "html-lsp", "css-lsp", "json-lsp",
        },
        run_on_start = not prebaked, -- in prod image, do NOT run at startup
        auto_update = false,
      }
    end,
  },

  -- snippets
  {
    "L3MON4D3/LuaSnip",
    build = "make install_jsregexp",
    config = function() require("luasnip.loaders.from_vscode").lazy_load() end,
    dependencies = { "rafamadriz/friendly-snippets" },
  },

  -- completion: nvim-cmp (+ cmdline source)
  { "hrsh7th/cmp-cmdline" },
  {
    "hrsh7th/nvim-cmp",
    event = "InsertEnter",
    dependencies = {
      "hrsh7th/cmp-nvim-lsp",
      "hrsh7th/cmp-buffer",
      "hrsh7th/cmp-path",
      "saadparwaiz1/cmp_luasnip",
    },
    opts = function()
      local cmp = require("cmp")
      local luasnip = require("luasnip")
      return {
        snippet = { expand = function(args) luasnip.lsp_expand(args.body) end },
        completion = { completeopt = "menu,menuone,noselect" },
        mapping = cmp.mapping.preset.insert({
          ["<CR>"]      = cmp.mapping.confirm({ select = true }),
          ["<C-Space>"] = cmp.mapping.complete(),
          ["<Tab>"]     = cmp.mapping(function(fallback)
            if cmp.visible() then
              cmp.select_next_item()
            elseif luasnip.expand_or_jumpable() then
              luasnip.expand_or_jump()
            else
              fallback()
            end
          end, { "i", "s" }),
          ["<S-Tab>"]   = cmp.mapping(function(fallback)
            if cmp.visible() then
              cmp.select_prev_item()
            elseif luasnip.jumpable(-1) then
              luasnip.jump(-1)
            else
              fallback()
            end
          end, { "i", "s" }),
          ["<Esc>"]     = function(fallback) cmp.abort(); fallback() end,
        }),
        sources = cmp.config.sources({
          { name = "nvim_lsp" },
          { name = "luasnip" },
          { name = "path" },
          { name = "buffer" },
        }),
        preselect = require("cmp").PreselectMode.Item,
      }
    end,
    config = function(_, opts)
      local cmp = require("cmp")
      cmp.setup(opts)

      -- cmdline completion ----------------------------------------------------
      -- ":" -> commands + paths
      cmp.setup.cmdline(":", {
        mapping = cmp.mapping.preset.cmdline(),
        sources = cmp.config.sources(
          { { name = "path" } },
          { { name = "cmdline" } }
        ),
      })
      -- "/" and "?" -> buffer words
      cmp.setup.cmdline({ "/", "?" }, {
        mapping = cmp.mapping.preset.cmdline(),
        sources = { { name = "buffer" } },
      })
    end,
  },

  -- autopairs (auto-close () {} "" etc.) + cmp integration
  {
    "windwp/nvim-autopairs",
    event = "InsertEnter",
    opts = {
      check_ts = true,
      fast_wrap = {}, -- <M-e> to wrap selection (optional)
    },
    config = function(_, opts)
      local npairs = require("nvim-autopairs")
      npairs.setup(opts)
      local ok, cmp = pcall(require, "cmp")
      if ok then
        local cmp_autopairs = require("nvim-autopairs.completion.cmp")
        cmp.event:on("confirm_done", cmp_autopairs.on_confirm_done())
      end
    end,
  },

  -- auto-close/rename HTML (and jsx/tsx if TS parsers installed)
  { "windwp/nvim-ts-autotag", event = "InsertEnter", opts = {} },

  -- LSP setup
  {
    "neovim/nvim-lspconfig",
    config = function()
      local lsp = require("lspconfig")
      local caps = require("cmp_nvim_lsp").default_capabilities()

      vim.diagnostic.config({ virtual_text = false, float = { border = "rounded" } })

      local function setup(server, conf)
        conf = conf or {}; conf.capabilities = caps; lsp[server].setup(conf)
      end

      setup("gopls")
      if lsp.ts_ls then setup("ts_ls") else setup("tsserver") end
      setup("html"); setup("cssls"); setup("jsonls")
      setup("lua_ls", {
        settings = {
          Lua = { workspace = { checkThirdParty = false }, diagnostics = { globals = { "vim" } } },
        },
      })

      local function map(m, lhs, rhs, d)
        vim.keymap.set(m, lhs, rhs, vim.tbl_extend("force", { silent = true }, { desc = d or "" }))
      end
      map("n", "gd", vim.lsp.buf.definition, "LSP definition")
      map("n", "gr", vim.lsp.buf.references, "LSP references")
      map("n", "K",  vim.lsp.buf.hover, "LSP hover")
      map("n", "<leader>rn", vim.lsp.buf.rename, "LSP rename")
      map("n", "<leader>ca", vim.lsp.buf.code_action, "LSP code action")
      map("n", "[d", vim.diagnostic.goto_prev, "Prev diagnostic")
      map("n", "]d", vim.diagnostic.goto_next, "Next diagnostic")
    end,
  },

  -- QoL
  { "lewis6991/gitsigns.nvim",   opts = {},                                     config = true },
  { "nvim-lualine/lualine.nvim", opts = { options = { theme = "tokyonight" } }, config = true },
  { "folke/which-key.nvim",      opts = {},                                     config = true },
  { "numToStr/Comment.nvim",     opts = {},                                     config = true },
})

-- telescope keys -------------------------------------------------------------
local ok_telescope, telescope = pcall(require, "telescope.builtin")
if ok_telescope then
  vim.keymap.set("n", "<leader>ff", telescope.find_files, { desc = "Find files" })
  vim.keymap.set("n", "<leader>fg", telescope.live_grep,  { desc = "Live grep" })
  vim.keymap.set("n", "<leader>fb", telescope.buffers,    { desc = "Buffers" })
  vim.keymap.set("n", "<leader>fh", telescope.help_tags,  { desc = "Help" })
end

-- exit current file (close buffer) -------------------------------------------
vim.keymap.set("n", "<leader>e", "<cmd>bd<CR>", { desc = "Close buffer" })

