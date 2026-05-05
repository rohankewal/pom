package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "v0.1.0"

type phase int

const (
	workPhase phase = iota
	breakPendingPhase // auto-start-break = false: waiting for user to begin break
	breakPhase
	donePhase
)

type viewMode int

const (
	timerView viewMode = iota
	statsView
	heatmapView
	promptView
)

type tickMsg time.Time

type model struct {
	// timer
	progress    progress.Model
	phase       phase
	workTime    time.Duration
	breakTime   time.Duration
	elapsed     time.Duration
	paused      bool
	activeBreak time.Duration
	isLongBreak bool
	// session
	title             string
	tags              []string
	rounds            int
	longBreak         time.Duration
	longBreakInterval int
	currentRound      int
	roundsCompleted   int
	interruptions     int
	startTime         time.Time
	sessions          []Session
	sessionSaved      bool
	// daily goal
	goalDuration time.Duration
	goalBar      progress.Model
	todayBase    time.Duration
	// UI
	cfg        appConfig
	viewMode   viewMode
	termWidth  int
	statsTable table.Model
	tableNotes []string
	noteInput  textinput.Model
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{tick()}
	if m.goalDuration > 0 {
		cmds = append(cmds, m.goalBar.SetPercent(m.goalPercent()))
	}
	return tea.Batch(cmds...)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) goalPercent() float64 {
	if m.goalDuration == 0 {
		return 0
	}
	focus := m.todayBase + time.Duration(m.roundsCompleted)*m.workTime
	pct := float64(focus) / float64(m.goalDuration)
	if pct > 1 {
		pct = 1
	}
	return pct
}

func (m model) toSession() Session {
	return Session{
		Date:            m.startTime,
		Title:           m.title,
		Tags:            m.tags,
		RoundsPlanned:   m.rounds,
		RoundsCompleted: m.roundsCompleted,
		WorkDuration:    m.workTime,
		Interruptions:   m.interruptions,
		Completed:       m.roundsCompleted == m.rounds,
	}
}

func (m model) breakDurationFor(roundsCompleted int) (time.Duration, bool) {
	if m.longBreakInterval > 0 && m.longBreak > 0 && roundsCompleted%m.longBreakInterval == 0 {
		return m.longBreak, true
	}
	return m.breakTime, false
}

func (m model) fmtRemaining(d time.Duration) string {
	if !m.cfg.ShowSeconds {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return d.Round(time.Second).String()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.viewMode == promptView {
		return m.updatePrompt(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			if !m.sessionSaved && m.roundsCompleted > 0 {
				m.sessions = append(m.sessions, m.toSession())
				saveSessions(m.sessions)
			}
			return m, tea.Quit
		}

		switch m.viewMode {
		case statsView:
			switch msg.String() {
			case "s", "esc":
				m.viewMode = timerView
			case "h":
				m.viewMode = heatmapView
			default:
				var cmd tea.Cmd
				m.statsTable, cmd = m.statsTable.Update(msg)
				return m, cmd
			}
			return m, nil

		case heatmapView:
			if msg.String() == "h" || msg.String() == "s" || msg.Type == tea.KeyEsc {
				m.viewMode = timerView
			}
			return m, nil

		default: // timerView
			switch msg.String() {
			case " ":
				if m.phase == breakPendingPhase {
					m.phase = breakPhase
					runHook(m.cfg.HookBreakStart, hookEnv(m))
					startMusic(m.cfg)
					return m, tea.Batch(tick(), m.progress.SetPercent(0))
				}
				m.paused = !m.paused
			case "n":
				return m.skipPhase()
			case "s":
				m.statsTable = buildStatsTable(m.sessions)
				m.tableNotes = buildTableNotes(m.sessions)
				m.viewMode = statsView
			case "h":
				m.viewMode = heatmapView
			case "i":
				m.interruptions++
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.progress.Width = msg.Width - 8
		m.goalBar.Width = msg.Width - 8

	case tickMsg:
		if m.paused || m.phase == breakPendingPhase {
			return m, tick()
		}
		m.elapsed += time.Second

		var total time.Duration
		if m.phase == workPhase {
			total = m.workTime
		} else {
			total = m.activeBreak
		}

		if m.elapsed >= total {
			if m.phase == workPhase {
				m.roundsCompleted++
				dur, isLong := m.breakDurationFor(m.roundsCompleted)
				m.activeBreak = dur
				m.isLongBreak = isLong
				m.paused = false
				m.elapsed = 0

				breakLabel := "take a break"
				if isLong {
					breakLabel = "long break time"
				}
				sendNotification("go-pom", fmt.Sprintf("Round %d done — %s!", m.roundsCompleted, breakLabel))
				runHook(m.cfg.HookWorkDone, hookEnv(m))
				stopMusic()

				if !m.cfg.AutoStartBreak {
					m.phase = breakPendingPhase
					return m, tea.Batch(m.progress.SetPercent(1), m.goalBar.SetPercent(m.goalPercent()))
				}
				m.phase = breakPhase
				runHook(m.cfg.HookBreakStart, hookEnv(m))
				return m, tea.Batch(tick(), m.progress.SetPercent(0), m.goalBar.SetPercent(m.goalPercent()))
			}

			// Break finished.
			if m.currentRound < m.rounds {
				m.currentRound++
				m.phase = workPhase
				m.paused = false
				m.elapsed = 0
				sendNotification("go-pom", fmt.Sprintf("Break over — round %d starting!", m.currentRound))
				runHook(m.cfg.HookBreakDone, hookEnv(m))
				runHook(m.cfg.HookWorkStart, hookEnv(m))
				startMusic(m.cfg)
				return m, tea.Batch(tick(), m.progress.SetPercent(0))
			}

			// All rounds done.
			if !m.sessionSaved {
				m.sessions = append(m.sessions, m.toSession())
				saveSessions(m.sessions)
				m.sessionSaved = true
			}
			sendNotification("go-pom", "All rounds complete! Great work.")
			runHook(m.cfg.HookSessionComplete, hookEnv(m))
			stopMusic()
			m.phase = donePhase
			m.viewMode = promptView
			return m, tea.Batch(m.progress.SetPercent(1), m.noteInput.Focus())
		}

		pct := float64(m.elapsed) / float64(total)
		if m.cfg.StatusFile {
			go writeStateFile(m)
		}
		return m, tea.Batch(tick(), m.progress.SetPercent(pct))

	case progress.FrameMsg:
		pm, cmd1 := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		gm, cmd2 := m.goalBar.Update(msg)
		m.goalBar = gm.(progress.Model)
		return m, tea.Batch(cmd1, cmd2)
	}

	return m, nil
}

func (m model) updatePrompt(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if note := m.noteInput.Value(); note != "" && m.sessionSaved && len(m.sessions) > 0 {
				m.sessions[len(m.sessions)-1].Note = note
				saveSessions(m.sessions)
			}
			return m, tea.Quit
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.noteInput, cmd = m.noteInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.viewMode {
	case statsView:
		return renderStats(computeStats(m.sessions), m.statsTable, m.tableNotes)
	case heatmapView:
		return renderHeatmap(m.sessions)
	case promptView:
		return fmt.Sprintf(
			"\n  Session complete! All %d rounds done.\n\n  What did you accomplish?\n  %s\n\n  %s\n",
			m.rounds,
			m.noteInput.View(),
			dimStyle.Render("Enter to save  •  Esc to skip"),
		)
	}

	if m.phase == donePhase {
		return "\n  All rounds complete! Great work.\n\n"
	}

	if m.phase == breakPendingPhase {
		return m.breakPendingView()
	}

	if m.cfg.Compact {
		return m.compactTimerView()
	}
	return m.fullTimerView()
}

func (m model) breakPendingView() string {
	breakLabel := "short break"
	if m.isLongBreak {
		breakLabel = "long break"
	}
	content := fmt.Sprintf(
		"\n  Round %d complete!\n\n  Your %s is ready: %s\n\n  %s\n\n  %s\n",
		m.roundsCompleted,
		breakLabel,
		fmtDur(m.activeBreak),
		accentStyle.Render("Press space to begin your break"),
		dimStyle.Render("n skip break  •  s stats  •  h heatmap  •  q quit"),
	)
	if m.cfg.Border && m.termWidth > 0 {
		return makeBorderStyle(m.termWidth).Render(content)
	}
	return content
}

func (m model) fullTimerView() string {
	label, total := m.phaseLabel()
	remaining := (total - m.elapsed).Round(time.Second)
	if remaining < 0 {
		remaining = 0
	}

	intrStr := m.intrStr()
	pauseStr := ""
	if m.paused {
		pauseStr = "  •  " + pausedStyle.Render("⏸ PAUSED")
	}

	out := fmt.Sprintf(
		"\n  %s — %s  (%s remaining)\n  Block %d of %d  •  %s%s\n\n  %s\n",
		label, m.title, m.fmtRemaining(remaining),
		m.currentRound, m.rounds, intrStr, pauseStr,
		m.progress.View(),
	)

	if m.goalDuration > 0 {
		focus := m.todayBase + time.Duration(m.roundsCompleted)*m.workTime
		out += fmt.Sprintf("\n  Daily Goal: %s / %s\n  %s\n",
			fmtDur(focus), fmtDur(m.goalDuration),
			m.goalBar.View(),
		)
	}

	spaceHint := "space pause"
	if m.paused {
		spaceHint = "space resume"
	}
	out += "\n  " + dimStyle.Render(spaceHint+"  •  n skip  •  i interrupt  •  s stats  •  h heatmap  •  q quit") + "\n"

	if m.cfg.Border && m.termWidth > 0 {
		return makeBorderStyle(m.termWidth).Render(out)
	}
	return out
}

func (m model) compactTimerView() string {
	label, total := m.phaseLabel()
	remaining := (total - m.elapsed).Round(time.Second)
	if remaining < 0 {
		remaining = 0
	}

	pauseStr := ""
	if m.paused {
		pauseStr = "  " + pausedStyle.Render("⏸")
	}

	spaceHint := "space pause"
	if m.paused {
		spaceHint = "space resume"
	}

	out := fmt.Sprintf(
		"\n  %s — %s  %s  •  Block %d/%d  •  %s%s\n  %s  %s\n",
		label, m.title, m.fmtRemaining(remaining),
		m.currentRound, m.rounds, m.intrStr(), pauseStr,
		m.progress.View(),
		dimStyle.Render(spaceHint+"  •  n  •  i  •  s  •  h  •  q"),
	)

	if m.goalDuration > 0 {
		focus := m.todayBase + time.Duration(m.roundsCompleted)*m.workTime
		out += fmt.Sprintf("  %s  %s / %s\n",
			m.goalBar.View(), fmtDur(focus), fmtDur(m.goalDuration))
	}

	if m.cfg.Border && m.termWidth > 0 {
		return makeBorderStyle(m.termWidth).Render(out)
	}
	return out
}

func (m model) skipPhase() (tea.Model, tea.Cmd) {
	switch m.phase {
	case workPhase:
		m.roundsCompleted++
		dur, isLong := m.breakDurationFor(m.roundsCompleted)
		m.activeBreak = dur
		m.isLongBreak = isLong
		m.elapsed = 0
		runHook(m.cfg.HookWorkDone, hookEnv(m))
		stopMusic()
		if !m.cfg.AutoStartBreak {
			m.phase = breakPendingPhase
			return m, tea.Batch(m.progress.SetPercent(1), m.goalBar.SetPercent(m.goalPercent()))
		}
		m.phase = breakPhase
		runHook(m.cfg.HookBreakStart, hookEnv(m))
		return m, tea.Batch(tick(), m.progress.SetPercent(0), m.goalBar.SetPercent(m.goalPercent()))
	case breakPhase, breakPendingPhase:
		if m.currentRound < m.rounds {
			m.currentRound++
			m.phase = workPhase
			m.elapsed = 0
			runHook(m.cfg.HookBreakDone, hookEnv(m))
			runHook(m.cfg.HookWorkStart, hookEnv(m))
			startMusic(m.cfg)
			return m, tea.Batch(tick(), m.progress.SetPercent(0))
		}
		if !m.sessionSaved {
			m.sessions = append(m.sessions, m.toSession())
			saveSessions(m.sessions)
			m.sessionSaved = true
		}
		runHook(m.cfg.HookSessionComplete, hookEnv(m))
		stopMusic()
		m.phase = donePhase
		m.viewMode = promptView
		return m, tea.Batch(m.progress.SetPercent(1), m.noteInput.Focus())
	}
	return m, nil
}

func (m model) phaseLabel() (string, time.Duration) {
	if m.phase == workPhase {
		return "Work", m.workTime
	}
	if m.isLongBreak {
		return "Long Break", m.activeBreak
	}
	return "Break", m.activeBreak
}

func (m model) intrStr() string {
	switch m.interruptions {
	case 0:
		return "no interruptions"
	case 1:
		return "1 interruption"
	default:
		return fmt.Sprintf("%d interruptions", m.interruptions)
	}
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "export":
			exportCSV()
			return
		case "status":
			format := "%p  %d  %r"
			fs := flag.NewFlagSet("status", flag.ExitOnError)
			fs.StringVar(&format, "format", format, "Output format: %p phase  %r remaining  %R minutes  %d title  %b block  %P pause-icon  %t tags  %% literal")
			_ = fs.Parse(os.Args[2:])
			printStatus(format)
			return
		}
	}

	writeDefaultConfig()
	cfg := loadConfig()
	applyTheme(cfg.Theme)

	bellEnabled = cfg.Bell
	notificationsEnabled = cfg.Notifications

	ver := flag.Bool("version", false, "Print version and exit")
	workTime := flag.Duration("work", cfg.Work, "Work block duration (e.g. 25m, 1h)")
	breakTime := flag.Duration("break", cfg.Break, "Short break duration")
	longBreak := flag.Duration("long-break", cfg.LongBreak, "Long break duration")
	longBreakInterval := flag.Int("long-break-interval", cfg.LongBreakInterval, "Long break after every N rounds (0 = disable)")
	sessionTitle := flag.String("title", cfg.Title, "Title for this session")
	sessionTags := flag.String("tags", "", "Comma-separated tags for this session (e.g. work,deepwork)")
	rounds := flag.Int("rounds", cfg.Rounds, "Number of work/break rounds")
	goal := flag.Duration("goal", cfg.Goal, "Daily focus goal (e.g. 2h). 0 = no goal")
	noAltScreen := flag.Bool("no-alt-screen", false, "Disable alternate screen (useful for recording demos)")
	flag.Parse()

	if *ver {
		fmt.Println("go-pom", version)
		return
	}

	sessions := loadSessions()

	todayBase := time.Duration(0)
	today := time.Now().Format("2006-01-02")
	for _, s := range sessions {
		if s.Date.Format("2006-01-02") == today {
			todayBase += time.Duration(s.RoundsCompleted) * s.WorkDuration
		}
	}

	ti := textinput.New()
	ti.Placeholder = "What did you accomplish this session?"
	ti.CharLimit = 120
	ti.Width = 52

	var tags []string
	if *sessionTags != "" {
		for _, t := range strings.Split(*sessionTags, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tags = append(tags, t)
			}
		}
	}

	m := model{
		progress:          buildProgressBar(cfg.ProgressStyle),
		goalBar:           buildGoalBar(),
		cfg:               cfg,
		phase:             workPhase,
		workTime:          *workTime,
		breakTime:         *breakTime,
		longBreak:         *longBreak,
		longBreakInterval: *longBreakInterval,
		activeBreak:       *breakTime,
		title:             *sessionTitle,
		tags:              tags,
		rounds:            *rounds,
		currentRound:      1,
		goalDuration:      *goal,
		todayBase:         todayBase,
		startTime:         time.Now(),
		sessions:          sessions,
		noteInput:         ti,
		termWidth:         80,
	}

	defer stopMusic()
	startMusic(cfg)
	runHook(cfg.HookWorkStart, hookEnv(m))

	opts := []tea.ProgramOption{}
	if !*noAltScreen {
		opts = append(opts, tea.WithAltScreen())
	}
	if _, err := tea.NewProgram(m, opts...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
