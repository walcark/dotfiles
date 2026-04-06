local M = {}

function M.compile_puml()
  local file = vim.api.nvim_buf_get_name(0)

  -- ensure .puml
  if file == '' or file:match '%s.puml$' then
    return
  end

  local cmd = { 'plantuml', '-tpng', '-o', './png', file }

  vim.fn.jobstart(cmd, {
    stdout_buffered = true,
    stderr_buffered = true,

    on_exit = function(_, code)
      if code == 0 then
        vim.notify('PlantUML compiled ✔', vim.log.levels.INFO, { title = 'PlantUML' })
      else
        vim.notify('PlantUML compilation failed ✖', vim.log.levels.ERROR, { title = 'PlantUML' })
      end
    end,
  })
end

return M
