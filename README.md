<div align="center">

```
██████╗  ██████╗ ███╗   ███╗
██╔══██╗██╔═══██╗████╗ ████║
██████╔╝██║   ██║██╔████╔██║
██╔═══╝ ██║   ██║██║╚██╔╝██║
██║     ╚██████╔╝██║ ╚═╝ ██║
╚═╝      ╚═════╝ ╚═╝     ╚═╝
```

**A focused Pomodoro timer that lives in your terminal.**

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/yourusername/pom?style=flat)](https://github.com/yourusername/pom/releases)

</div>

---

`pom` is a terminal Pomodoro timer with analytics, themes, lo-fi music, lifecycle hooks, and tmux status-line integration — everything you need to build a deep-work habit, in a single binary with no runtime dependencies.

> **Demo GIF coming soon** — run `pom` to see it live.

---

## Features

- **Pomodoro timer** — configurable work blocks, short breaks, and long breaks
- **Analytics** — productivity score, 7-day bar chart, 15-week heatmap, streak tracking
- **14 built-in themes** — default, Dracula, Catppuccin (all four), Nord, Gruvbox, Solarized, Tokyo Night, One Dark, Rose Pine, Kanagawa
- **Daily focus goal** — visual progress bar tracks cumulative focus across sessions
- **Lo-fi music** — streams curated SomaFM stations during work blocks (requires mpv or ffplay)
- **Lifecycle hooks** — run shell commands when phases start, end, or complete
- **tmux / status-line integration** — write live state to a JSON file; query it from anywhere
- **Session tags** — tag sessions and filter analytics by project or context
- **Interruption tracking** — log distractions with a single keypress
- **Session notes** — capture what you accomplished at the end of every session
- **CSV export** — pipe session history to your spreadsheet or data pipeline
- **Zero dependencies** — single statically-linked binary, works on macOS and Linux

---

## Install

### macOS / Linux — one-liner

```sh
curl -fsSL https://raw.githubusercontent.com/rohankewal/pom/main/install.sh | sh
```

### Homebrew

```sh
brew install yourusername/tap/pom
```

### Go

```sh
go install github.com/rohankewal/pom@latest
```

### Download a binary

Pre-built binaries for macOS (arm64/amd64), Linux (arm64/amd64), and Windows are on the [Releases page](https://github.com/rohankewal/pom/releases).

---

## Quick Start

```sh
# Start a default session (25m work × 4 rounds)
pom

# Name your session and tag it
pom --title "API redesign" --tags "work,backend"

# Custom durations
pom --work 50m --break 10m --rounds 3

# Set a daily focus goal
pom --goal 4h
```

That's it. `pom` creates `~/.config/pom/config` on first run — edit it to make your preferences permanent.

---

## Keybindings

| Key | Action |
|-----|--------|
| `space` | Pause / resume (or begin break when `auto-start-break = false`) |
| `n` | Skip to next phase |
| `i` | Log an interruption |
| `s` | Open analytics |
| `h` | Open focus heatmap |
| `q` / `ctrl+c` | Quit (auto-saves the session) |

**In analytics:**

| Key | Action |
|-----|--------|
| `↑` / `↓` | Scroll session history |
| `h` | Switch to heatmap |
| `s` / `esc` | Return to timer |

---

## Configuration

`pom` writes a self-documenting config to `~/.config/pom/config` on first launch. Every option can also be overridden with a flag at runtime.

```ini
# ~/.config/pom/config

# ── Timer ──────────────────────────────────────────────────────────────
work                = 25m
break               = 5m
long-break          = 15m
rounds              = 4
long-break-interval = 4    # long break after every N rounds (0 = off)
title               = new session
goal                = 0    # daily focus goal, e.g. 4h

# ── Appearance ─────────────────────────────────────────────────────────
theme               = default
border              = false
progress-style      = gradient   # gradient | solid

# ── Behaviour ──────────────────────────────────────────────────────────
bell                = true
notifications       = true
show-seconds        = true
auto-start-break    = true   # false = press space to start your break
compact             = false

# ── Music ──────────────────────────────────────────────────────────────
music               = false       # requires mpv, ffplay, vlc, or mplayer
music-player        =             # leave blank to auto-detect
music-volume        = 50

# ── Lifecycle Hooks ────────────────────────────────────────────────────
# hook-work-start       = terminal-notifier -message "Focus time, $GOPOM_TITLE"
# hook-work-done        = osascript -e 'display notification "Round done"'
# hook-break-start      =
# hook-break-done       =
# hook-session-complete = say "Great work today"

# ── Status Line ────────────────────────────────────────────────────────
status-file = false   # writes ~/.pom/state.json every second
```

### All CLI flags

```
--work              Work block duration        (default: 25m)
--break             Short break duration       (default: 5m)
--long-break        Long break duration        (default: 15m)
--long-break-interval  Long break every N rounds (default: 4)
--rounds            Number of rounds           (default: 4)
--title             Session title
--tags              Comma-separated tags       (e.g. work,backend)
--goal              Daily focus goal           (e.g. 4h)
--version           Print version and exit
```

---

## Analytics

Press `s` at any time to open the analytics panel.

```
  Analytics

  Score  87/100 — Great work

  Focus Time  3h 25m        Sessions  14      Completion  85%
  Streak      6 days        Best      12      Avg Interruptions  0.8

  Last 7 Days
  ────────────────────────────────────────
  Mon  ████████████████░░░░  1h 25m
  Tue  ████████████████████  2h 0m
  Wed  ████████████░░░░░░░░  1h 0m
  Thu  ████████░░░░░░░░░░░░  45m
  Fri  ████████████████████  2h 0m
  Sat  ░░░░░░░░░░░░░░░░░░░░
  Sun  ████░░░░░░░░░░░░░░░░  25m

  Recent Sessions
  ...
```

Press `h` to see a **15-week heatmap** of your focus history.

### Productivity Score

Your score (0–100) is calculated from three signals:

| Signal | Weight |
|--------|--------|
| Session completion rate | 40% |
| Current streak (normalised to 21 days) | 30% |
| Focus quality (inverse of avg interruptions) | 30% |

---

## Themes

Switch theme in config or at the command line:

```sh
pom --theme tokyo-night    # one-off override not available yet — set in config
```

| Theme | Description |
|-------|-------------|
| `default` | Purple gradient |
| `dracula` | Classic Dracula palette |
| `catppuccin-mocha` | Mocha — dark |
| `catppuccin-latte` | Latte — light |
| `catppuccin-frappe` | Frappé |
| `catppuccin-macchiato` | Macchiato |
| `nord` | Arctic, north-bluish |
| `gruvbox` | Warm retro orange |
| `solarized-dark` | Solarized dark |
| `solarized-light` | Solarized light |
| `tokyo-night` | Tokyo Night blue |
| `one-dark` | Atom One Dark |
| `rose-pine` | Rosé Pine |
| `kanagawa` | Kanagawa Wave |

---

## Lo-fi Music

`pom` plays a random SomaFM lo-fi stream during work blocks and pauses during breaks. It auto-detects a media player — install one if you don't have it:

```sh
# macOS
brew install mpv

# Linux
apt install mpv   # or ffmpeg for ffplay
```

Then enable in config:

```ini
music        = true
music-volume = 60
```

Streams cycle through: Groove Salad · Fluid · Lush · Drone Zone · Suburbs of Goa · Space Station

---

## Hooks

Run any shell command when a phase starts or ends. The following environment variables are available to every hook:

| Variable | Value |
|----------|-------|
| `GOPOM_TITLE` | Session title |
| `GOPOM_TAGS` | Comma-separated tags |
| `GOPOM_PHASE` | `work` / `break` / `done` |
| `GOPOM_ROUND` | Current round number |
| `GOPOM_TOTAL_ROUNDS` | Total rounds planned |
| `GOPOM_INTERRUPTIONS` | Interruptions so far |
| `GOPOM_REMAINING` | Time remaining (e.g. `24m30s`) |

**Example hooks:**

```ini
# macOS notification when work starts
hook-work-start = osascript -e 'display notification "Focus!" with title "pom"'

# Slack status via API when on a break
hook-break-start = curl -s -X POST $SLACK_API ...

# Log every session to a file
hook-session-complete = echo "$(date): $GOPOM_TITLE [$GOPOM_TAGS]" >> ~/focus-log.txt

# Trigger a webhook (Zapier, Make, etc.)
hook-session-complete = curl -s -X POST https://hooks.zapier.com/... \
    -d "{\"title\":\"$GOPOM_TITLE\",\"tags\":\"$GOPOM_TAGS\"}"
```

---

## tmux & Status Line Integration

Enable the state file in config:

```ini
status-file = true
```

`pom` writes `~/.pom/state.json` every second. Read it with the `status` subcommand:

```sh
pom status                         # work  25m0s  API redesign
pom status --format "%p %b %r"     # work 1/4 24m30s
pom status --format "[%d %b %P]"   # [API redesign 1/4 ⏸]
```

**Format placeholders:**

| Placeholder | Output |
|-------------|--------|
| `%p` | Phase (`work` / `break`) |
| `%r` | Remaining (e.g. `24m30s`) |
| `%R` | Remaining in minutes (e.g. `24m`) |
| `%d` | Session title |
| `%b` | Block (`1/4`) |
| `%P` | Pause icon (`⏸` or empty) |
| `%t` | Tags (comma-separated) |
| `%%` | Literal `%` |

**tmux status bar** (`~/.tmux.conf`):

```
set -g status-right '#(pom status --format "[%d %b %r %P]")'
set -g status-interval 1
```

**Starship** (`~/.config/starship.toml`):

```toml
[custom.pom]
command = "pom status --format '%p %b %r'"
when = "pom status 2>/dev/null"
format = "[$output]($style) "
style = "bold purple"
```

---

## Exporting Data

Export all sessions to CSV:

```sh
pom export > sessions.csv
```

**Columns:** Date, Title, Rounds Completed, Rounds Planned, Work Duration (min), Focus Time (min), Interruptions, Completed, Note

Session history is stored as plain JSON at `~/.pom/sessions.json` — readable, portable, and yours.

---

## Data & Privacy

All data lives locally on your machine:

| Path | Contents |
|------|----------|
| `~/.config/pom/config` | Your configuration |
| `~/.pom/sessions.json` | Session history |
| `~/.pom/state.json` | Live timer state (when `status-file = true`) |

`pom` never phones home. No telemetry, no accounts, no cloud.

---

## Building from Source

```sh
git clone https://github.com/yourusername/pom
cd pom
go build -o pom .
./pom
```

Requires Go 1.24+.

---

## Contributing

Pull requests are welcome. For significant changes, open an issue first to discuss what you'd like to change.

```sh
git clone https://github.com/yourusername/pom
cd pom
go build ./...
go vet ./...
```

---

## License

MIT © Rohan Kewalramani(https://github.com/rohankewal)

---

<div align="center">

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) · [Bubbles](https://github.com/charmbracelet/bubbles) · [Lip Gloss](https://github.com/charmbracelet/lipgloss)

</div>
