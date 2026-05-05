package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type appConfig struct {
	Work              time.Duration
	Break             time.Duration
	LongBreak         time.Duration
	Rounds            int
	LongBreakInterval int
	Title             string
	Goal              time.Duration
	// Visual
	Theme         string
	Border        bool
	ProgressStyle string // "gradient" or "solid"
	// Behaviour
	Bell           bool
	Notifications  bool
	ShowSeconds    bool
	AutoStartBreak bool
	Compact        bool
	// Music
	Music       bool
	MusicPlayer string
	MusicVolume int
	// Lifecycle hooks (shell commands)
	HookWorkStart       string
	HookWorkDone        string
	HookBreakStart      string
	HookBreakDone       string
	HookSessionComplete string
	// tmux / status-line integration
	StatusFile bool
}

func defaultConfig() appConfig {
	return appConfig{
		Work:              25 * time.Minute,
		Break:             5 * time.Minute,
		LongBreak:         15 * time.Minute,
		Rounds:            4,
		LongBreakInterval: 4,
		Title:             "new session",
		Theme:             "default",
		Border:            false,
		ProgressStyle:     "gradient",
		Bell:              true,
		Notifications:     true,
		ShowSeconds:       true,
		AutoStartBreak:    true,
		Compact:           false,
		Music:             false,
		MusicPlayer:       "",
		MusicVolume:       50,
		StatusFile:        false,
	}
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "go-pom", "config")
}

func parseBool(s string) (bool, bool) {
	switch strings.ToLower(s) {
	case "true", "yes", "1":
		return true, true
	case "false", "no", "0":
		return false, true
	}
	return false, false
}

func loadConfig() appConfig {
	cfg := defaultConfig()
	f, err := os.Open(configPath())
	if err != nil {
		return cfg
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "work":
			if d, err := time.ParseDuration(val); err == nil {
				cfg.Work = d
			}
		case "break":
			if d, err := time.ParseDuration(val); err == nil {
				cfg.Break = d
			}
		case "long-break":
			if d, err := time.ParseDuration(val); err == nil {
				cfg.LongBreak = d
			}
		case "rounds":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.Rounds = n
			}
		case "long-break-interval":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.LongBreakInterval = n
			}
		case "title":
			cfg.Title = val
		case "goal":
			if val != "0" && val != "" {
				if d, err := time.ParseDuration(val); err == nil {
					cfg.Goal = d
				}
			}
		case "theme":
			cfg.Theme = val
		case "border":
			if b, ok := parseBool(val); ok {
				cfg.Border = b
			}
		case "progress-style":
			if val == "solid" || val == "gradient" {
				cfg.ProgressStyle = val
			}
		case "bell":
			if b, ok := parseBool(val); ok {
				cfg.Bell = b
			}
		case "notifications":
			if b, ok := parseBool(val); ok {
				cfg.Notifications = b
			}
		case "show-seconds":
			if b, ok := parseBool(val); ok {
				cfg.ShowSeconds = b
			}
		case "auto-start-break":
			if b, ok := parseBool(val); ok {
				cfg.AutoStartBreak = b
			}
		case "compact":
			if b, ok := parseBool(val); ok {
				cfg.Compact = b
			}
		case "music":
			if b, ok := parseBool(val); ok {
				cfg.Music = b
			}
		case "music-player":
			cfg.MusicPlayer = val
		case "music-volume":
			if n, err := strconv.Atoi(val); err == nil && n >= 0 && n <= 100 {
				cfg.MusicVolume = n
			}
		case "hook-work-start":
			cfg.HookWorkStart = val
		case "hook-work-done":
			cfg.HookWorkDone = val
		case "hook-break-start":
			cfg.HookBreakStart = val
		case "hook-break-done":
			cfg.HookBreakDone = val
		case "hook-session-complete":
			cfg.HookSessionComplete = val
		case "status-file":
			if b, ok := parseBool(val); ok {
				cfg.StatusFile = b
			}
		}
	}
	return cfg
}

func writeDefaultConfig() {
	path := configPath()
	if _, err := os.Stat(path); err == nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, []byte(`# go-pom configuration
# ~/.config/go-pom/config
#
# Command-line flags override these values.
# Durations use Go format: 25m, 1h, 1h30m

# ── Timer ─────────────────────────────────────────────────────────────────────

# Duration of each work block
work = 25m

# Duration of each short break
break = 5m

# Duration of the long break (taken every long-break-interval rounds)
long-break = 15m

# Number of rounds per session
rounds = 4

# Take a long break after this many work blocks (0 = disable)
long-break-interval = 4

# Default session title
title = new session

# Daily focus goal — e.g. 2h, 90m (0 = no goal)
goal = 0

# ── Appearance ────────────────────────────────────────────────────────────────

# Color theme. Available themes:
#   default, dracula, catppuccin-mocha, catppuccin-latte, catppuccin-frappe,
#   catppuccin-macchiato, nord, gruvbox, solarized-dark, solarized-light,
#   tokyo-night, one-dark, rose-pine, kanagawa
theme = default

# Draw a rounded border around the timer
border = false

# Progress bar style: gradient or solid
progress-style = gradient

# ── Behaviour ─────────────────────────────────────────────────────────────────

# Play a terminal bell when a phase ends
bell = true

# Send an OS system notification when a phase ends
notifications = true

# Show seconds in the remaining-time countdown
show-seconds = true

# Start the break timer automatically when work ends.
# Set to false to let you choose when the break begins.
auto-start-break = true

# Compact single-pane layout (fewer blank lines, shorter hints)
compact = false

# ── Music ─────────────────────────────────────────────────────────────────────

# Play lo-fi background music during work blocks (requires mpv, ffplay, vlc, or mplayer)
music = false

# Preferred player binary name. Leave blank to auto-detect.
music-player =

# Playback volume 0–100
music-volume = 50

# ── Lifecycle Hooks ───────────────────────────────────────────────────────────
# Shell commands run at phase transitions.
# Available env vars: GOPOM_TITLE, GOPOM_TAGS, GOPOM_PHASE, GOPOM_ROUND,
#                     GOPOM_TOTAL_ROUNDS, GOPOM_INTERRUPTIONS, GOPOM_REMAINING

# hook-work-start =
# hook-work-done =
# hook-break-start =
# hook-break-done =
# hook-session-complete =

# ── Status Line ───────────────────────────────────────────────────────────────

# Write ~/.go-pom/state.json every tick (use with: go-pom status --format "...")
# Placeholders: %p phase  %r remaining  %R minutes  %d title  %b block  %P pause-icon  %t tags  %% literal
status-file = false
`), 0644)
}
