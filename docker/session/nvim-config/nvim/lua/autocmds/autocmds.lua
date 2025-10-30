-- Create an augroup to hold our autocmds, which is good practice
vim.o.updatetime = 200
local diagnostics_group = vim.api.nvim_create_augroup("Diagnostics", { clear = true })

-- This autocmd runs whenever the cursor stops moving in normal mode
vim.api.nvim_create_autocmd("CursorHold", {
  group = diagnostics_group,
  pattern = "*",
  callback = function()
    -- Get diagnostics for the current line
    local diagnostics = vim.diagnostic.get(0, { lnum = vim.fn.line('.') - 1 })

    -- If there are any diagnostics on this line, open the float
    if #diagnostics > 0 then
      vim.diagnostic.open_float(nil, {
        focusable = false, -- Prevents the float from stealing focus
        scope = "line",
      })
    end
  end,
  desc = "Automatically show diagnostics on cursor hold",
})
