require("config.lazy")
require("config.options")
require("mappings.remaps")
require("autocmds.autocmds")
vim.cmd('let $LANG = "en_US.UTF-8"')
if vim.loader then
  vim.loader.enable()
end


-- Color scheme
vim.cmd([[colorscheme zenbones]])
-- vim.cmd([[colorscheme nord]])
-- vim.cmd([[colorscheme kanagawa]])
-- Remove background for normal and floating windows
vim.api.nvim_set_hl(0, "Normal", { bg = "none" })
vim.api.nvim_set_hl(0, "NormalFloat", { bg = "none" })

-- Remove background for Telescope windows
vim.api.nvim_set_hl(0, "TelescopeNormal", { bg = "none" })
vim.api.nvim_set_hl(0, "TelescopeBorder", { bg = "none" })
vim.api.nvim_set_hl(0, "TelescopePromptNormal", { bg = "none" })
vim.api.nvim_set_hl(0, "TelescopeResultsNormal", { bg = "none" })
vim.api.nvim_set_hl(0, "TelescopePreviewNormal", { bg = "none" })
