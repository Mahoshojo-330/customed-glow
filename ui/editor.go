package ui

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/editor"
)

// defaultEditor is used when $EDITOR is not set. x/editor would fall back
// to nano; we prefer vim, so we set $EDITOR ourselves before deferring to
// editor.Cmd.
const defaultEditor = "vim"

type editorFinishedMsg struct{ err error }

func editorCmd(path string, lineno int) (*exec.Cmd, error) {
	if strings.TrimSpace(os.Getenv("EDITOR")) == "" {
		// x/editor falls back to nano when $EDITOR is unset; we prefer vim.
		os.Setenv("EDITOR", defaultEditor)
	}
	return editor.Cmd("Glow", path, editor.LineNumber(uint(lineno))) //nolint:gosec
}

func openEditor(path string, lineno int) tea.Cmd {
	cb := func(err error) tea.Msg {
		return editorFinishedMsg{err}
	}
	cmd, err := editorCmd(path, lineno)
	if err != nil {
		return func() tea.Msg { return cb(err) }
	}
	return tea.ExecProcess(cmd, cb)
}
