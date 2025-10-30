vim.keymap.set("n", "<leader>e", ":Ex<CR>", {
  desc = "Open file explorer (netrw)",
  noremap = true,
  silent = true,
})

vim.keymap.set("v", "J", ":m '>+1<CR>gv=gv")
vim.keymap.set("v", "K", ":m '<-2<CR>gv=gv")

vim.keymap.set("n", "<leader>sp", ":StickyPad")
vim.keymap.set("n", "<leader>su", ":Unfold")
vim.keymap.set("n", "<leader>sf", ":Fold")

-- vim.keymap.set('n', '<leader>se', vim.diagnostic.open_float, { desc = "[S]how [E]rror diagnostics" })
