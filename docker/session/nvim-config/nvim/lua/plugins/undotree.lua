-- lua/plugins/undotree.lua
return {
  {
    "mbbill/undotree",
    init = function()
      local undodir = vim.fn.stdpath("data") .. "/undodir"
      if vim.fn.isdirectory(undodir) == 0 then
        vim.fn.mkdir(undodir, "p")
      end
      vim.opt.undodir = undodir

      vim.opt.undofile = true
    end,

    config = function()
      vim.keymap.set("n", "<leader>u", vim.cmd.UndotreeToggle, { desc = "Toggle UndoTree" })
    end,
  },
}
