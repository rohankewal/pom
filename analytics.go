package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Session struct {
	Date            time.Time     `json:"date"`
	Title           string        `json:"title"`
	Tags            []string      `json:"tags,omitempty"`
	RoundsPlanned   int           `json:"rounds_planned"`
	RoundsCompleted int           `json:"rounds_completed"`
	WorkDuration    time.Duration `json:"work_duration"`
	Interruptions   int           `json:"interruptions"`
	Note            string        `json:"note,omitempty"`
	Completed       bool          `json:"completed"`
}

func dataPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".go-pom", "sessions.json")
}

func loadSessions() []Session {
	data, err := os.ReadFile(dataPath())
	if err != nil {
		return nil
	}
	var out []Session
	_ = json.Unmarshal(data, &out)
	return out
}

func saveSessions(sessions []Session) {
	path := dataPath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.MarshalIndent(sessions, "", "  ")
	_ = os.WriteFile(path, data, 0644)
}

func exportCSV() {
	sessions := loadSessions()
	if len(sessions) == 0 {
		fmt.Fprintln(os.Stderr, "no sessions recorded yet")
		return
	}
	w := csv.NewWriter(os.Stdout)
	_ = w.Write([]string{
		"Date", "Title", "Rounds Completed", "Rounds Planned",
		"Work Duration (min)", "Focus Time (min)", "Interruptions", "Completed", "Note",
	})
	for _, s := range sessions {
		focus := time.Duration(s.RoundsCompleted) * s.WorkDuration
		_ = w.Write([]string{
			s.Date.Format("2006-01-02 15:04:05"),
			s.Title,
			strconv.Itoa(s.RoundsCompleted),
			strconv.Itoa(s.RoundsPlanned),
			strconv.Itoa(int(s.WorkDuration.Minutes())),
			strconv.Itoa(int(focus.Minutes())),
			strconv.Itoa(s.Interruptions),
			strconv.FormatBool(s.Completed),
			s.Note,
		})
	}
	w.Flush()
}

type analyticsStats struct {
	totalSessions     int
	totalFocusTime    time.Duration
	completionRate    float64
	currentStreak     int
	bestStreak        int
	avgInterruptions  float64
	productivityScore int
	last7Days         [7]time.Duration
}

func computeStats(sessions []Session) analyticsStats {
	var s analyticsStats
	s.totalSessions = len(sessions)
	if len(sessions) == 0 {
		return s
	}

	dailyFocus := make(map[string]time.Duration)
	completed, totalInterruptions := 0, 0
	for _, sess := range sessions {
		focus := time.Duration(sess.RoundsCompleted) * sess.WorkDuration
		s.totalFocusTime += focus
		dailyFocus[sess.Date.Format("2006-01-02")] += focus
		if sess.Completed {
			completed++
		}
		totalInterruptions += sess.Interruptions
	}
	s.completionRate = float64(completed) / float64(s.totalSessions) * 100
	s.avgInterruptions = float64(totalInterruptions) / float64(s.totalSessions)

	now := time.Now()
	for i := 0; i < 7; i++ {
		s.last7Days[i] = dailyFocus[now.AddDate(0, 0, -(6-i)).Format("2006-01-02")]
	}

	days := make([]string, 0, len(dailyFocus))
	for d := range dailyFocus {
		days = append(days, d)
	}
	sort.Strings(days)

	daySet := make(map[string]bool, len(days))
	for _, d := range days {
		daySet[d] = true
	}

	start := now.Format("2006-01-02")
	if !daySet[start] {
		start = now.AddDate(0, 0, -1).Format("2006-01-02")
	}
	for d := start; daySet[d]; {
		s.currentStreak++
		t, _ := time.Parse("2006-01-02", d)
		d = t.AddDate(0, 0, -1).Format("2006-01-02")
	}

	run := 0
	for i, day := range days {
		if i == 0 {
			run = 1
		} else {
			prev, _ := time.Parse("2006-01-02", days[i-1])
			curr, _ := time.Parse("2006-01-02", day)
			if curr.Sub(prev) == 24*time.Hour {
				run++
			} else {
				run = 1
			}
		}
		if run > s.bestStreak {
			s.bestStreak = run
		}
	}

	streakFrac := float64(s.currentStreak) / 21.0
	if streakFrac > 1 {
		streakFrac = 1
	}
	intrFrac := s.avgInterruptions / 5.0
	if intrFrac > 1 {
		intrFrac = 1
	}
	s.productivityScore = int(s.completionRate*0.4 + streakFrac*30 + (1-intrFrac)*30)

	return s
}

func scoreLabel(score int) string {
	switch {
	case score >= 91:
		return "In the zone"
	case score >= 81:
		return "Great work"
	case score >= 61:
		return "Solid focus"
	case score >= 41:
		return "Building habits"
	default:
		return "Getting started"
	}
}

func fmtDur(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

const barMaxWidth = 20

func buildTableNotes(sessions []Session) []string {
	notes := make([]string, 0)
	for i := len(sessions) - 1; i >= 0 && len(notes) < 20; i-- {
		notes = append(notes, sessions[i].Note)
	}
	return notes
}

func renderStats(stats analyticsStats, t table.Model, notes []string) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  " + accentStyle.Render("Analytics") + "\n\n")

	scoreStr := fmt.Sprintf("%d/100 — %s", stats.productivityScore, scoreLabel(stats.productivityScore))
	b.WriteString("  Score  " + scoreStyle.Render(scoreStr) + "\n\n")

	b.WriteString(fmt.Sprintf("  Focus Time  %-12s  Sessions  %-6d  Completion  %.0f%%\n",
		fmtDur(stats.totalFocusTime), stats.totalSessions, stats.completionRate))
	b.WriteString(fmt.Sprintf("  Streak      %-12s  Best      %-10d  Avg Interruptions  %.1f\n\n",
		fmt.Sprintf("%d days", stats.currentStreak), stats.bestStreak, stats.avgInterruptions))

	b.WriteString("  " + accentStyle.Render("Last 7 Days") + "\n")
	b.WriteString("  " + dimStyle.Render(strings.Repeat("─", 40)) + "\n")

	maxDur := time.Minute
	for _, d := range stats.last7Days {
		if d > maxDur {
			maxDur = d
		}
	}
	now := time.Now()
	for i, dur := range stats.last7Days {
		label := now.AddDate(0, 0, -(6-i)).Format("Mon")
		barLen := int(float64(dur) / float64(maxDur) * barMaxWidth)
		var bar, pad string
		if barLen > 0 {
			bar = barStyle.Render(strings.Repeat("█", barLen))
			pad = strings.Repeat(" ", barMaxWidth-barLen)
		} else {
			bar = dimStyle.Render("─")
			pad = strings.Repeat(" ", barMaxWidth-1)
		}
		durStr := ""
		if dur > 0 {
			durStr = fmtDur(dur)
		}
		b.WriteString(fmt.Sprintf("  %-4s  %s%s  %s\n", label, bar, pad, dimStyle.Render(durStr)))
	}

	b.WriteString("\n")
	b.WriteString("  " + accentStyle.Render("Recent Sessions") + "\n")
	for _, line := range strings.Split(t.View(), "\n") {
		b.WriteString("  " + line + "\n")
	}

	cursor := t.Cursor()
	if cursor >= 0 && cursor < len(notes) && notes[cursor] != "" {
		b.WriteString("\n  " + dimStyle.Render("Note: ") + notes[cursor] + "\n")
	} else {
		b.WriteString("\n")
	}

	b.WriteString("\n  " + dimStyle.Render("↑/↓ scroll  •  h heatmap  •  s return  •  q quit") + "\n")
	return b.String()
}

func buildStatsTable(sessions []Session) table.Model {
	cols := []table.Column{
		{Title: "Date", Width: 10},
		{Title: "Title", Width: 14},
		{Title: "Tags", Width: 12},
		{Title: "Rounds", Width: 7},
		{Title: "Focus", Width: 7},
		{Title: "Int", Width: 4},
		{Title: "✓", Width: 2},
	}

	rows := make([]table.Row, 0, len(sessions))
	for i := len(sessions) - 1; i >= 0 && len(rows) < 20; i-- {
		s := sessions[i]
		focus := time.Duration(s.RoundsCompleted) * s.WorkDuration
		done := "✗"
		if s.Completed {
			done = "✓"
		}
		tags := strings.Join(s.Tags, ",")
		rows = append(rows, table.Row{
			s.Date.Format("2006-01-02"),
			s.Title,
			tags,
			fmt.Sprintf("%d/%d", s.RoundsCompleted, s.RoundsPlanned),
			fmtDur(focus),
			strconv.Itoa(s.Interruptions),
			done,
		})
	}

	ts := table.DefaultStyles()
	ts.Header = ts.Header.Bold(true).Foreground(lipgloss.Color(activeTheme.TableHdr))
	ts.Selected = ts.Selected.
		Foreground(lipgloss.Color(activeTheme.TableSelFg)).
		Background(lipgloss.Color(activeTheme.TableSelBg))

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)
	t.SetStyles(ts)
	return t
}

func heatCell(dur time.Duration, future bool) string {
	if future {
		return "  "
	}
	switch {
	case dur == 0:
		return heatLevels[0].Render("██")
	case dur < 30*time.Minute:
		return heatLevels[1].Render("██")
	case dur < time.Hour:
		return heatLevels[2].Render("██")
	case dur < 2*time.Hour:
		return heatLevels[3].Render("██")
	default:
		return heatLevels[4].Render("██")
	}
}

func renderHeatmap(sessions []Session) string {
	dailyFocus := make(map[string]time.Duration)
	for _, s := range sessions {
		dailyFocus[s.Date.Format("2006-01-02")] += time.Duration(s.RoundsCompleted) * s.WorkDuration
	}

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	start := monday.AddDate(0, 0, -14*7)

	const weeks = 15
	dayLabels := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	grid := [7][weeks]string{}

	for w := 0; w < weeks; w++ {
		for d := 0; d < 7; d++ {
			date := start.AddDate(0, 0, w*7+d)
			grid[d][w] = heatCell(dailyFocus[date.Format("2006-01-02")], date.After(now))
		}
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  " + accentStyle.Render("Focus Heatmap") + " — last 15 weeks\n")
	b.WriteString("  " + dimStyle.Render(strings.Repeat("─", 52)) + "\n\n")

	for d := 0; d < 7; d++ {
		b.WriteString(fmt.Sprintf("  %-4s  ", dayLabels[d]))
		for w := 0; w < weeks; w++ {
			b.WriteString(grid[d][w])
			if w < weeks-1 {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n  " + dimStyle.Render("Less "))
	for i, lvl := range heatLevels {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(lvl.Render("██"))
	}
	b.WriteString(dimStyle.Render(" More"))
	b.WriteString("\n\n  " + dimStyle.Render("h or s return  •  q quit") + "\n")
	return b.String()
}
