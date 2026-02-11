return {
  "obsidian-nvim/obsidian.nvim",
  -- the obsidian vault in this default config  ~/obsidian-vault
  -- If you want to use the home shortcut '~' here you need to call 'vim.fn.expand':
  -- event = { "bufreadpre " .. vim.fn.expand "~" .. "/my-vault/**.md" },
  event = {
    "BufReadPre " .. vim.fn.expand("~/PARA/ressources/notes/obsidian-vault") .. "*.md",
    "BufNewFile " .. vim.fn.expand("~/PARA/ressources/notes/obsidian-vault") .. "*.md",
  },

  dependencies = {
    "nvim-lua/plenary.nvim",
    { "hrsh7th/nvim-cmp", optional = true },
    {
      "AstroNvim/astrocore",
      opts = {
        mappings = {
          n = {
            ["gf"] = {
              function()
                local ok, obsidian = pcall(require, "obsidian")
                local has_util = ok and obsidian.util ~= nil
                local has_fn = has_util and type(obsidian.util.cursor_on_markdown_link) == "function"

                if has_fn and obsidian.util.cursor_on_markdown_link() then
                  return "<Cmd>ObsidianFollowLink<CR>"
                end

                -- fallback Vim natif
                return "gf"
              end,
              desc = "Obsidian Follow Link",
              expr = true,
              silent = true,
            },
          },
        },
      },
    },
  },
  opts = function(_, opts)
    local astrocore = require "astrocore"
    return astrocore.extend_tbl(opts, {
      workspaces = {
        {
          path = "~/PARA/ressources/notes/obsidian-vault", -- specify the vault location. no need to call 'vim.fn.expand' here
        },
      },
      open = {
        use_advanced_uri = true,
      },
      finder = (astrocore.is_available "snacks.pick" and "snacks.pick")
        or (astrocore.is_available "telescope.nvim" and "telescope.nvim")
        or (astrocore.is_available "fzf-lua" and "fzf-lua")
        or (astrocore.is_available "mini.pick" and "mini.pick"),

      templates = {
        subdir = "templates",
        date_format = "%Y-%m-%d-%a",
        time_format = "%H:%M",
      },
      daily_notes = {
        folder = "daily",
      },
      completion = {
        nvim_cmp = astrocore.is_available "nvim-cmp",
        blink = astrocore.is_available "blink",
      },

      note_frontmatter_func = function(note)
        -- This is equivalent to the default frontmatter function.
        local out = { id = note.id, aliases = note.aliases, tags = note.tags }
        -- `note.metadata` contains any manually added fields in the frontmatter.
        -- So here we just make sure those fields are kept in the frontmatter.
        if note.metadata ~= nil and not vim.tbl_isempty(note.metadata) then
          for k, v in pairs(note.metadata) do
            out[k] = v
          end
        end
        return out
      end,

      -- Optional, by default when you use `:ObsidianFollowLink` on a link to an external
      -- URL it will be ignored but you can customize this behavior here.
      follow_url_func = vim.ui.open,
    })
  end,
}
