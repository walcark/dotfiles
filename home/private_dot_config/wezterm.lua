local wezterm = require 'wezterm'
local act = wezterm.action
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
-- Split window -- 
config.keys = {
	{ 
		key = "h",
		mods = "WIN",
		action = act.splitpane({
			direction = "left",
			size = { Percent = 50 },
		}),
	},
	{ 
		key = "j",
		mods = "WIN",
		action = act.splitpane({
			direction = "down",
			size = { Percent = 50 },
		}),
	},
	{ 
		key = "k",
		mods = "WIN",
		action = act.splitpane({
			direction = "up",
			size = { Percent = 50 },
		}),
	},
	{ 
		key = "l",
		mods = "WIN",
		action = act.splitpane({
			direction = "right",
			size = { Percent = 50 },
		}),
	},
}

-- Shell --
config.default_prog = {"/bin/bash"}

return config
