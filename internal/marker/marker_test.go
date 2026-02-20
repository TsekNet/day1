package marker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMarkerRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T)
		wantExists  bool
		wantWriteOK bool
	}{
		{
			name:        "fresh directory has no marker",
			setup:       func(t *testing.T) {},
			wantExists:  false,
			wantWriteOK: true,
		},
		{
			name: "marker exists after write",
			setup: func(t *testing.T) {
				t.Helper()
				if err := Write(); err != nil {
					t.Fatalf("Write: %v", err)
				}
			},
			wantExists:  true,
			wantWriteOK: true,
		},
		{
			name: "marker gone after remove",
			setup: func(t *testing.T) {
				t.Helper()
				if err := Write(); err != nil {
					t.Fatalf("Write: %v", err)
				}
				if err := Remove(); err != nil {
					t.Fatalf("Remove: %v", err)
				}
			},
			wantExists:  false,
			wantWriteOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			setConfigHome(t, tmp)

			tt.setup(t)

			got, err := Exists()
			if err != nil {
				t.Fatalf("Exists: %v", err)
			}
			if got != tt.wantExists {
				t.Errorf("Exists() = %v, want %v", got, tt.wantExists)
			}

			if tt.wantWriteOK {
				if err := Write(); err != nil {
					t.Fatalf("Write: %v", err)
				}
				exists, err := Exists()
				if err != nil {
					t.Fatalf("Exists after Write: %v", err)
				}
				if !exists {
					t.Error("expected marker to exist after Write")
				}
			}
		})
	}
}

func TestWriteCreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	setConfigHome(t, tmp)

	if err := Write(); err != nil {
		t.Fatalf("Write: %v", err)
	}

	p := filepath.Join(tmp, appDir, fileName)
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("sentinel not found at %s: %v", p, err)
	}
	if info.Size() == 0 {
		t.Error("sentinel file is empty, expected timestamp content")
	}
}

func TestRemoveNonexistent(t *testing.T) {
	tmp := t.TempDir()
	setConfigHome(t, tmp)

	if err := Remove(); err != nil {
		t.Fatalf("Remove on nonexistent file should not error, got: %v", err)
	}
}

// setConfigHome points the config directory env vars at tmp so tests don't
// touch the real user config. Cannot be used with t.Parallel().
func setConfigHome(t *testing.T, tmp string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	if os.PathSeparator == '\\' {
		t.Setenv("LOCALAPPDATA", tmp)
	}
}
