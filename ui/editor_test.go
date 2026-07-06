package ui

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestEditorCmd(t *testing.T) {
	tests := []struct {
		name      string
		editorEnv string
		want      []string
	}{
		{
			name:      "falls back to vim when EDITOR is unset",
			editorEnv: "",
			want:      []string{"vim", "+12", "foo.md"},
		},
		{
			name:      "respects EDITOR when set",
			editorEnv: "nano",
			want:      []string{"nano", "+12", "foo.md"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("EDITOR", tc.editorEnv)

			cmd, err := editorCmd("foo.md", 12)
			if err != nil {
				t.Fatalf("editorCmd returned error: %v", err)
			}

			if filepath.Base(cmd.Args[0]) != tc.want[0] {
				t.Errorf("Args[0] = %q, want %q", cmd.Args[0], tc.want[0])
			}
			if !slices.Equal(cmd.Args[1:], tc.want[1:]) {
				t.Errorf("Args[1:] = %v, want %v", cmd.Args[1:], tc.want[1:])
			}
		})
	}
}
