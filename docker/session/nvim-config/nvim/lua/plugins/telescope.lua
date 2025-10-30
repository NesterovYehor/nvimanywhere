return {
  {
    'nvim-telescope/telescope.nvim',
    tag = '0.1.8',
    dependencies = {
      'nvim-lua/plenary.nvim',
    },
    config = function()
      vim.keymap.set("n", "<leader>ff", require('telescope.builtin').find_files)
      vim.keymap.set("n", "<leader>fg", require('telescope.builtin').live_grep)
      vim.keymap.set('n', '<leader>fh', '<cmd>Telescope help_tags<CR>', {
        desc = "Find help tags"
      })
      require("telescope").setup({
        defaults = {
          file_ignore_patterns = {
            "pythonenv/",
            "%.pyc",
            "__pycache__/",
            "%.venv/",
            "vendor",
            "bin",
          },
        },
        extensions = {
          ["ui-select"] = {
            require("telescope.themes").get_dropdown({}),
          },
        },
      })
    end
  },
  {
    "nvim-telescope/telescope-ui-select.nvim",
    config = function()
      require("telescope").setup({
        extensions = {
          ["ui-select"] = {
            require("telescope.themes").get_dropdown({
            }),
          },
        },
      })
      require("telescope").load_extension("ui-select")
    end,
  },
}
