package main

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// Theme defines every color used throughout the TUI.
type Theme struct {
	Accent     string // headers, progress bar, bar chart
	Dim        string // secondary / hint text
	Score      string // productivity score highlight
	Paused     string // ⏸ PAUSED indicator
	GoalFill   string // goal bar solid fill
	GradStart  string // work/break progress bar gradient start (hex)
	GradEnd    string // work/break progress bar gradient end   (hex)
	Heat       [5]string // heatmap intensity levels 0–4
	Border     string // rounded border color
	TableHdr   string // sessions table header
	TableSelFg string // selected row foreground
	TableSelBg string // selected row background
}

var builtinThemes = map[string]Theme{
	"default": {
		Accent:     "#875fff",
		Dim:        "#626262",
		Score:      "#af87ff",
		Paused:     "#ffaf00",
		GoalFill:   "#d75fd7",
		GradStart:  "#875fff",
		GradEnd:    "#d75fd7",
		Heat:       [5]string{"#3a3a3a", "#5f00ff", "#875fff", "#af5fff", "#af87ff"},
		Border:     "#626262",
		TableHdr:   "#875fff",
		TableSelFg: "#ffffaf",
		TableSelBg: "#5f00ff",
	},
	"dracula": {
		Accent:     "#bd93f9",
		Dim:        "#6272a4",
		Score:      "#50fa7b",
		Paused:     "#ffb86c",
		GoalFill:   "#ff79c6",
		GradStart:  "#bd93f9",
		GradEnd:    "#ff79c6",
		Heat:       [5]string{"#44475a", "#6272a4", "#bd93f9", "#8be9fd", "#50fa7b"},
		Border:     "#6272a4",
		TableHdr:   "#bd93f9",
		TableSelFg: "#f8f8f2",
		TableSelBg: "#6272a4",
	},
	"catppuccin-mocha": {
		Accent:     "#cba6f7",
		Dim:        "#585b70",
		Score:      "#a6e3a1",
		Paused:     "#fab387",
		GoalFill:   "#f38ba8",
		GradStart:  "#cba6f7",
		GradEnd:    "#89b4fa",
		Heat:       [5]string{"#313244", "#45475a", "#cba6f7", "#89b4fa", "#a6e3a1"},
		Border:     "#585b70",
		TableHdr:   "#cba6f7",
		TableSelFg: "#cdd6f4",
		TableSelBg: "#45475a",
	},
	"catppuccin-latte": {
		Accent:     "#8839ef",
		Dim:        "#8c8fa1",
		Score:      "#40a02b",
		Paused:     "#fe640b",
		GoalFill:   "#ea76cb",
		GradStart:  "#8839ef",
		GradEnd:    "#04a5e5",
		Heat:       [5]string{"#dce0e8", "#bcc0cc", "#8839ef", "#04a5e5", "#40a02b"},
		Border:     "#bcc0cc",
		TableHdr:   "#8839ef",
		TableSelFg: "#4c4f69",
		TableSelBg: "#e6e9ef",
	},
	"catppuccin-frappe": {
		Accent:     "#ca9ee6",
		Dim:        "#626880",
		Score:      "#a6d189",
		Paused:     "#ef9f76",
		GoalFill:   "#f4b8e4",
		GradStart:  "#ca9ee6",
		GradEnd:    "#8caaee",
		Heat:       [5]string{"#303446", "#51576d", "#ca9ee6", "#8caaee", "#a6d189"},
		Border:     "#626880",
		TableHdr:   "#ca9ee6",
		TableSelFg: "#c6d0f5",
		TableSelBg: "#414559",
	},
	"catppuccin-macchiato": {
		Accent:     "#c6a0f6",
		Dim:        "#6e738d",
		Score:      "#a6da95",
		Paused:     "#f5a97f",
		GoalFill:   "#f5bde6",
		GradStart:  "#c6a0f6",
		GradEnd:    "#8aadf4",
		Heat:       [5]string{"#1e2030", "#5b6078", "#c6a0f6", "#8aadf4", "#a6da95"},
		Border:     "#6e738d",
		TableHdr:   "#c6a0f6",
		TableSelFg: "#cad3f5",
		TableSelBg: "#363a4f",
	},
	"nord": {
		Accent:     "#88c0d0",
		Dim:        "#4c566a",
		Score:      "#a3be8c",
		Paused:     "#d08770",
		GoalFill:   "#81a1c1",
		GradStart:  "#88c0d0",
		GradEnd:    "#81a1c1",
		Heat:       [5]string{"#2e3440", "#3b4252", "#88c0d0", "#81a1c1", "#a3be8c"},
		Border:     "#4c566a",
		TableHdr:   "#88c0d0",
		TableSelFg: "#eceff4",
		TableSelBg: "#3b4252",
	},
	"gruvbox": {
		Accent:     "#fe8019",
		Dim:        "#928374",
		Score:      "#b8bb26",
		Paused:     "#fb4934",
		GoalFill:   "#fabd2f",
		GradStart:  "#fe8019",
		GradEnd:    "#fabd2f",
		Heat:       [5]string{"#282828", "#504945", "#fe8019", "#fabd2f", "#b8bb26"},
		Border:     "#504945",
		TableHdr:   "#fe8019",
		TableSelFg: "#fbf1c7",
		TableSelBg: "#504945",
	},
	"solarized-dark": {
		Accent:     "#268bd2",
		Dim:        "#586e75",
		Score:      "#859900",
		Paused:     "#cb4b16",
		GoalFill:   "#2aa198",
		GradStart:  "#268bd2",
		GradEnd:    "#2aa198",
		Heat:       [5]string{"#002b36", "#073642", "#268bd2", "#2aa198", "#859900"},
		Border:     "#586e75",
		TableHdr:   "#268bd2",
		TableSelFg: "#fdf6e3",
		TableSelBg: "#073642",
	},
	"solarized-light": {
		Accent:     "#268bd2",
		Dim:        "#93a1a1",
		Score:      "#859900",
		Paused:     "#cb4b16",
		GoalFill:   "#2aa198",
		GradStart:  "#268bd2",
		GradEnd:    "#2aa198",
		Heat:       [5]string{"#eee8d5", "#93a1a1", "#268bd2", "#2aa198", "#859900"},
		Border:     "#93a1a1",
		TableHdr:   "#268bd2",
		TableSelFg: "#002b36",
		TableSelBg: "#eee8d5",
	},
	"tokyo-night": {
		Accent:     "#7aa2f7",
		Dim:        "#565f89",
		Score:      "#9ece6a",
		Paused:     "#ff9e64",
		GoalFill:   "#bb9af7",
		GradStart:  "#7aa2f7",
		GradEnd:    "#bb9af7",
		Heat:       [5]string{"#1a1b26", "#24283b", "#7aa2f7", "#bb9af7", "#9ece6a"},
		Border:     "#565f89",
		TableHdr:   "#7aa2f7",
		TableSelFg: "#c0caf5",
		TableSelBg: "#283457",
	},
	"one-dark": {
		Accent:     "#61afef",
		Dim:        "#5c6370",
		Score:      "#98c379",
		Paused:     "#e5c07b",
		GoalFill:   "#c678dd",
		GradStart:  "#61afef",
		GradEnd:    "#c678dd",
		Heat:       [5]string{"#282c34", "#3e4451", "#61afef", "#c678dd", "#98c379"},
		Border:     "#5c6370",
		TableHdr:   "#61afef",
		TableSelFg: "#abb2bf",
		TableSelBg: "#3e4451",
	},
	"rose-pine": {
		Accent:     "#c4a7e7",
		Dim:        "#6e6a86",
		Score:      "#9ccfd8",
		Paused:     "#f6c177",
		GoalFill:   "#ebbcba",
		GradStart:  "#c4a7e7",
		GradEnd:    "#ebbcba",
		Heat:       [5]string{"#191724", "#1f1d2e", "#c4a7e7", "#ebbcba", "#9ccfd8"},
		Border:     "#6e6a86",
		TableHdr:   "#c4a7e7",
		TableSelFg: "#e0def4",
		TableSelBg: "#26233a",
	},
	"kanagawa": {
		Accent:     "#7e9cd8",
		Dim:        "#727169",
		Score:      "#76946a",
		Paused:     "#ffa066",
		GoalFill:   "#957fb8",
		GradStart:  "#7e9cd8",
		GradEnd:    "#957fb8",
		Heat:       [5]string{"#1f1f28", "#2a2a37", "#7e9cd8", "#957fb8", "#76946a"},
		Border:     "#727169",
		TableHdr:   "#7e9cd8",
		TableSelFg: "#dcd7ba",
		TableSelBg: "#2a2a37",
	},
}

// Package-level style vars updated by applyTheme.
var (
	activeTheme Theme
	heatLevels  [5]lipgloss.Style

	accentStyle lipgloss.Style
	scoreStyle  lipgloss.Style
	pausedStyle lipgloss.Style
	dimStyle    lipgloss.Style
	barStyle    lipgloss.Style
)

func init() {
	applyTheme("default")
}

func applyTheme(name string) {
	t, ok := builtinThemes[name]
	if !ok {
		t = builtinThemes["default"]
	}
	activeTheme = t

	accentStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.Accent))
	scoreStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.Score))
	pausedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.Paused))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim))
	barStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent))

	for i, c := range t.Heat {
		heatLevels[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(c))
	}
}

func buildProgressBar(progressStyle string) progress.Model {
	if progressStyle == "solid" {
		return progress.New(progress.WithSolidFill(activeTheme.Accent))
	}
	return progress.New(progress.WithGradient(activeTheme.GradStart, activeTheme.GradEnd))
}

func buildGoalBar() progress.Model {
	return progress.New(progress.WithSolidFill(activeTheme.GoalFill))
}

func makeBorderStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(activeTheme.Border)).
		Padding(0, 1).
		Width(width - 4)
}
