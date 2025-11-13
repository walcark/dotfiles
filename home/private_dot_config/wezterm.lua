local wezterm = require 'wezterm'

local config = {}

-- Global appearance --
config.color_scheme = 'Tokyo Night Storm'
config.font = wezterm.font("JetBrainsMono Nerd Font")
config.font_size = 12.0

-- Window --
config.window_background_opacity = 0.95
config.window_decorations = "RESIZE"
config.enable_tab_bar = false
config.enable_scroll_bar = true
config.window_padding = {
    left = 5, 
	right = 5, 
	top = 5, 
	bottom = 5,
}

-- Shell --
config.default_prog = {"/bin/bash"}

return config
