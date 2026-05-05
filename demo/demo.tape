Output demo.gif

Require pom

Set Shell "zsh"
Set FontSize 15
Set FontFamily "JetBrainsMono NFM"
Set Width 1000
Set Height 550
Set Theme "Catppuccin Mocha"
Set Padding 20
Set PlaybackSpeed 1.0
Set TypingSpeed 60ms
Set CursorBlink false

# ── Launch ─────────────────────────────────────────────────────────────────────
Sleep 500ms
Type "pom --no-alt-screen --title 'deep work' --tags 'code,focus' --rounds 3 --work 10s --break 4s"
Sleep 400ms
Enter
Sleep 2s

# ── Watch the timer tick ───────────────────────────────────────────────────────
Sleep 6s

# ── Open analytics ─────────────────────────────────────────────────────────────
Type "s"
Sleep 3.5s

# ── Open heatmap ───────────────────────────────────────────────────────────────
Type "h"
Sleep 3.5s

# ── Back to timer ──────────────────────────────────────────────────────────────
Type "s"
Sleep 3s

# ── Quit ───────────────────────────────────────────────────────────────────────
Type "q"
Sleep 500ms
