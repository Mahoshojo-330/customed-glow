package ui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour/styles"
)

func TestGlamourRenderRendersMermaid(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper is a POSIX shell script")
	}

	dir := t.TempDir()
	command := filepath.Join(dir, "mermaid-ascii")
	if err := os.WriteFile(command, []byte("#!/bin/sh\nprintf '+---+\\n| A |\\n+---+\\n'\n"), 0o755); err != nil {
		t.Fatalf("create Mermaid renderer: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	previousConfig := config
	config.GlamourEnabled = true
	t.Cleanup(func() { config = previousConfig })

	m := pagerModel{
		common: &commonModel{cfg: Config{
			GlamourEnabled:   true,
			GlamourMaxWidth:  80,
			GlamourStyle:     styles.NoTTYStyle,
			PreserveNewLines: true,
			ShowLineNumbers:  false,
		}},
		viewport:        viewport.New(80, 24),
		currentDocument: markdown{Note: "diagram.md"},
	}

	rendered, err := glamourRender(m, "```mermaid\ngraph TD\n  A --> B\n```\n")
	if err != nil {
		t.Fatalf("glamourRender() error = %v", err)
	}
	if !strings.Contains(rendered, "| A |") {
		t.Errorf("rendered Mermaid graph missing: %q", rendered)
	}
	if strings.Contains(rendered, "graph TD") {
		t.Errorf("Mermaid source was not replaced: %q", rendered)
	}
}
