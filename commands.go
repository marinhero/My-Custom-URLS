package main

import (
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// clipboardMsg is returned after a clipboard operation
type clipboardMsg struct {
	success bool
	text    string
}

// copyToClipboardCmd copies text to the system clipboard
func copyToClipboardCmd(text string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("pbcopy")
		case "linux":
			cmd = exec.Command("xclip", "-selection", "clipboard")
		case "windows":
			cmd = exec.Command("clip")
		default:
			return clipboardMsg{success: false, text: text}
		}

		cmd.Stdin = strings.NewReader(text)
		err := cmd.Run()
		if err != nil {
			return clipboardMsg{success: false, text: text}
		}

		return clipboardMsg{success: true, text: text}
	}
}
