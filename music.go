package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"sync"
)

var lofiStreams = []string{
	"https://ice1.somafm.com/groovesalad-256-mp3",
	"https://ice1.somafm.com/fluid-128-mp3",
	"https://ice1.somafm.com/lush-128-mp3",
	"https://ice1.somafm.com/dronezone-256-mp3",
	"https://ice1.somafm.com/suburbsofgoa-128-mp3",
	"https://ice1.somafm.com/spacestation-128-mp3",
}

var (
	musicMu   sync.Mutex
	musicProc *exec.Cmd
)

func detectPlayer(preferred string) string {
	candidates := []string{preferred, "mpv", "ffplay", "vlc", "mplayer"}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if path, err := exec.LookPath(p); err == nil && path != "" {
			return p
		}
	}
	return ""
}

func playerArgs(player, url string, volume int) []string {
	switch player {
	case "mpv":
		return []string{"--no-video", "--quiet", fmt.Sprintf("--volume=%d", volume), url}
	case "ffplay":
		return []string{"-nodisp", "-loglevel", "quiet", "-volume", fmt.Sprintf("%d", volume), url}
	case "vlc":
		vlcVol := volume * 256 / 100
		return []string{"--intf", "dummy", "--quiet", "--no-video", fmt.Sprintf("--volume=%d", vlcVol), url}
	case "mplayer":
		return []string{"-nogui", "-really-quiet", "-volume", fmt.Sprintf("%d", volume), url}
	}
	return []string{url}
}

func startMusic(cfg appConfig) {
	if !cfg.Music {
		return
	}
	player := detectPlayer(cfg.MusicPlayer)
	if player == "" {
		return
	}
	stopMusic()
	url := lofiStreams[rand.Intn(len(lofiStreams))]
	args := playerArgs(player, url, cfg.MusicVolume)
	cmd := exec.Command(player, args...)
	musicMu.Lock()
	musicProc = cmd
	musicMu.Unlock()
	_ = cmd.Start()
}

func stopMusic() {
	musicMu.Lock()
	defer musicMu.Unlock()
	if musicProc != nil && musicProc.Process != nil {
		_ = musicProc.Process.Kill()
		musicProc = nil
	}
}
