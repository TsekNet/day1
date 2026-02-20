package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionSubcommand(t *testing.T) {
	root := buildRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("version command: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "day1") {
		t.Errorf("version output missing 'day1': %s", out)
	}
}

func TestFlagDefaults(t *testing.T) {
	root := buildRootCmd()

	tests := []struct {
		flag    string
		wantDef string
	}{
		{"pages-dir", ""},
		{"force", "false"},
		{"verbose", "false"},
	}

	for _, tt := range tests {
		f := root.Flags().Lookup(tt.flag)
		if f == nil {
			t.Fatalf("flag %q not found", tt.flag)
		}
		if f.DefValue != tt.wantDef {
			t.Errorf("flag %q default = %q, want %q", tt.flag, f.DefValue, tt.wantDef)
		}
	}
}

func TestRemovedFlags(t *testing.T) {
	root := buildRootCmd()
	removed := []string{"title", "help-url", "brand-name", "brand-logo", "final-page", "theme"}

	for _, name := range removed {
		if root.Flags().Lookup(name) != nil {
			t.Errorf("flag %q should not exist (moved to day1.yml)", name)
		}
	}
}

func TestInvalidPagesDir(t *testing.T) {
	root := buildRootCmd()
	root.SetArgs([]string{"--pages-dir", "/nonexistent/path/that/does/not/exist", "--force"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid pages-dir, got nil")
	}
	if !strings.Contains(err.Error(), "load pages") {
		t.Errorf("error = %q, want it to contain 'load pages'", err.Error())
	}
}
