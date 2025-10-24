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

  -- telescope (+ dependency)
  { "nvim-lua/plenary.nvim" },
  { "nvim-telescope/telescope.nvim" },

  -- treesitter
  {
    "nvim-treesitter/nvim-treesitter",
    build = ":TSUpdate",
    opts = {
      ensure_installed = { "lua", "go", "javascript", "html", "css" },
      highlight = { enable = true },
      indent = { enable = true },
    },
    config = function(_, opts)
      require("nvim-treesitter.configs").setup(opts)
    end,
  },

  -- completion: nvim-cmp (fast, no native deps)
  {
    "hrsh7th/nvim-cmp",
    dependencies = {
      "hrsh7th/cmp-nvim-lsp",
      "hrsh7th/cmp-buffer",
      "hrsh7th/cmp-path",
      -- optional snippet support:
      -- "L3MON4D3/LuaSnip", "saadparwaiz1/cmp_luasnip"
    },
    opts = function()
      local cmp = require("cmp")
      return {
        completion = { completeopt = "menu,menuone,noselect" },
        mapping = cmp.mapping.preset.insert({
          ["<CR>"]      = cmp.mapping.confirm({ select = true }),
          ["<C-Space>"] = cmp.mapping.complete(),
          ["<Tab>"]     = cmp.mapping(function(fallback)
            if cmp.visible() then cmp.select_next_item() else fallback() end
          end, { "i", "s" }),
          ["<S-Tab>"]   = cmp.mapping(function(fallback)
            if cmp.visible() then cmp.select_prev_item() else fallback() end
          end, { "i", "s" }),
          ["<Esc>"]     = function(fallback)
            cmp.abort(); fallback()
          end,
        }),
        sources = cmp.config.sources({
          { name = "nvim_lsp" },
          { name = "path" },
          { name = "buffer" },
        }),
        preselect = cmp.PreselectMode.Item,
      }
    end,
    config = function(_, opts)
      require("cmp").setup(opts)
    end,
  },

  -- LSP: add cmp capabilities
  {
    "neovim/nvim-lspconfig",
    config = function()
      local lsp = require("lspconfig")
      local caps = require("cmp_nvim_lsp").default_capabilities()
      vim.diagnostic.config({ virtual_text = false, float = { border = "rounded" } })

      local function setup(server, conf)
        conf = conf or {}
        conf.capabilities = caps
        lsp[server].setup(conf)
      end

      setup("gopls")
      setup("tsserver") -- if deprecated on your version, use "ts_ls"
      setup("html")
      setup("cssls")

      -- keymaps
      local function map(m, lhs, rhs, d)
        vim.keymap.set(m, lhs, rhs, vim.tbl_extend("force", { silent = true }, { desc = d or "" }))
      end
      map("n", "gd", vim.lsp.buf.definition, "LSP definition")
      map("n", "gr", vim.lsp.buf.references, "LSP references")
      map("n", "K", vim.lsp.buf.hover, "LSP hover")
      map("n", "<leader>rn", vim.lsp.buf.rename, "LSP rename")
      map("n", "<leader>ca", vim.lsp.buf.code_action, "LSP code action")
      map("n", "[d", vim.diagnostic.goto_prev, "Prev diagnostic")
      map("n", "]d", vim.diagnostic.goto_next, "Next diagnostic")
    end,
  },

  -- QoL (optional but nice)
  { "lewis6991/gitsigns.nvim",   opts = {},                                     config = true },
  { "nvim-lualine/lualine.nvim", opts = { options = { theme = "tokyonight" } }, config = true },
  { "folke/which-key.nvim",      opts = {},                                     config = true },
  { "numToStr/Comment.nvim",     opts = {},                                     config = true },
})

-- telescope minimal keymaps ---------------------------------------------------
local ok_telescope, telescope = pcall(require, "telescope.builtin")
if ok_telescope then
  vim.keymap.set("n", "<leader>ff", telescope.find_files, { desc = "Find files" })
  vim.keymap.set("n", "<leader>fg", telescope.live_grep, { desc = "Live grep" })
  vim.keymap.set("n", "<leader>fb", telescope.buffers, { desc = "Buffers" })
  vim.keymap.set("n", "<leader>fh", telescope.help_tags, { desc = "Help" })
end
