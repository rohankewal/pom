package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

var (
	bellEnabled          = true
	notificationsEnabled = true
)

func sendNotification(title, body string) {
	if bellEnabled {
		fmt.Print("\a")
	}
	if !notificationsEnabled {
		return
	}
	switch runtime.GOOS {
	case "darwin":
		exec.Command("osascript", "-e",
			fmt.Sprintf(`display notification %q with title %q`, body, title),
		).Run()
	case "linux":
		exec.Command("notify-send", title, body).Run()
	}
}
