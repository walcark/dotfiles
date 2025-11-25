require "nvchad.mappings"

-- add yours here

local map = vim.keymap.set
local default_opts = {noremap = true}

map("n", ";", ":", { desc = "CMD enter command mode" })
map("i", "jk", "<ESC>")
map('n', '<leader>ff', "<cmd>lua require'telescope.builtin'.find_files({ find_command = {'rg', '--files', '--hidden', '-g', '!.git' }})<cr>", default_opts)
-- map({ "n", "i", "v" }, "<C-s>", "<cmd> w <cr>")
