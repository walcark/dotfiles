vim.api.nvim_create_autocmd('BufWritePost', {
  pattern = '*.puml',
  callback = function()
    require('custom.plantuml').compile_puml()
  end,
})
