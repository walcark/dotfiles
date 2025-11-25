local wezterm = require 'wezterm'
local act = wezterm.action
local config = {}

-- Global appearance --
config.color_scheme = 'Tokyo Night Storm'
config.font = wezterm.font("JetBrainsMonoNL Nerd Font")
config.font_size = 12.0

-- Window --
config.window_background_opacity = 0.95
config.window_decorations = "RESIZE"
config.enable_tab_bar = false
config.harfbuzz_features = { 'calt=0' }
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
		mods = "CTRL",
		action = act.SplitPane({
			direction = "Left",
			size = { Percent = 50 },
		}),
	},
	{ 
		key = "j",
		mods = "CTRL",
		action = act.SplitPane({
			direction = "Down",
			size = { Percent = 50 },
		}),
	},
	{ 
		key = "k",
		mods = "CTRL",
		action = act.SplitPane({
			direction = "Up",
			size = { Percent = 50 },
		}),
	},
	{ 
		key = "l",
		mods = "CTRL",
		action = act.SplitPane({
			direction = "Right",
			size = { Percent = 50 },
		}),
	},
}

-- Shell --
config.default_prog = {"/bin/bash"}

return config
