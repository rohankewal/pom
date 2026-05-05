package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type timerState struct {
	Phase         string    `json:"phase"`
	Round         int       `json:"round"`
	TotalRounds   int       `json:"total_rounds"`
	Remaining     string    `json:"remaining"`
	RemainingM    int       `json:"remaining_minutes"`
	Paused        bool      `json:"paused"`
	Title         string    `json:"title"`
	Tags          []string  `json:"tags"`
	Interruptions int       `json:"interruptions"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func statePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".go-pom", "state.json")
}

func writeStateFile(m model) {
	phase := "work"
	switch m.phase {
	case breakPhase:
		phase = "break"
	case breakPendingPhase:
		phase = "break-pending"
	case donePhase:
		phase = "done"
	}
	_, total := m.phaseLabel()
	remaining := (total - m.elapsed).Round(time.Second)
	if remaining < 0 {
		remaining = 0
	}
	state := timerState{
		Phase:         phase,
		Round:         m.currentRound,
		TotalRounds:   m.rounds,
		Remaining:     remaining.String(),
		RemainingM:    int(remaining.Minutes()),
		Paused:        m.paused,
		Title:         m.title,
		Tags:          m.tags,
		Interruptions: m.interruptions,
		UpdatedAt:     time.Now(),
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	path := statePath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, data, 0644)
}

func hookEnv(m model) map[string]string {
	_, total := m.phaseLabel()
	remaining := (total - m.elapsed).Round(time.Second)
	if remaining < 0 {
		remaining = 0
	}
	phase := "work"
	switch m.phase {
	case breakPhase:
		phase = "break"
	case breakPendingPhase:
		phase = "break-pending"
	case donePhase:
		phase = "done"
	}
	return map[string]string{
		"GOPOM_TITLE":         m.title,
		"GOPOM_TAGS":          strings.Join(m.tags, ","),
		"GOPOM_PHASE":         phase,
		"GOPOM_ROUND":         fmt.Sprintf("%d", m.currentRound),
		"GOPOM_TOTAL_ROUNDS":  fmt.Sprintf("%d", m.rounds),
		"GOPOM_INTERRUPTIONS": fmt.Sprintf("%d", m.interruptions),
		"GOPOM_REMAINING":     remaining.String(),
	}
}

func runHook(command string, env map[string]string) {
	if command == "" {
		return
	}
	cmd := exec.Command("sh", "-c", command)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	go func() { _ = cmd.Run() }()
}

func printStatus(format string) {
	data, err := os.ReadFile(statePath())
	if err != nil {
		fmt.Println("no active session")
		return
	}
	var state timerState
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Println("error reading state")
		return
	}
	pauseIcon := ""
	if state.Paused {
		pauseIcon = "⏸"
	}
	out := strings.NewReplacer(
		"%p", state.Phase,
		"%r", state.Remaining,
		"%R", fmt.Sprintf("%dm", state.RemainingM),
		"%d", state.Title,
		"%b", fmt.Sprintf("%d/%d", state.Round, state.TotalRounds),
		"%P", pauseIcon,
		"%t", strings.Join(state.Tags, ","),
		"%%", "%",
	).Replace(format)
	fmt.Println(out)
}
