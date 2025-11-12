local wezterm = require 'wezterm'

return {
  -- --- Appearance ---
  color_scheme = "Builtin Solarized Dark",  -- theme intégré
  font = wezterm.font("JetBrainsMono Nerd Font"),  -- police lisible
  font_size = 11.0,

  -- --- Window ---
  window_background_opacity = 0.95,  -- légère transparence
  enable_tab_bar = false,            -- pas de barre d’onglets
  window_padding = {
    left = 5, right = 5, top = 5, bottom = 5,
  },

  -- --- Behavior ---
  default_prog = {"/bin/bash"},      -- ou "zsh", "fish", etc.
}
