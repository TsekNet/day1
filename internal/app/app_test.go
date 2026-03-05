package app

import (
	"os"
	"testing"

	"github.com/TsekNet/day1/internal/pages"
)

func testPages(n int) []pages.Page {
	pp := make([]pages.Page, n)
	for i := range pp {
		pp[i] = pages.Page{
			Frontmatter: pages.Frontmatter{Title: "Page " + string(rune('A'+i))},
			Markdown:    "# Page " + string(rune('A'+i)),
			SourceFile:  "page.md",
		}
	}
	return pp
}

func testApp(n int, cfg Config) *App { return New(testPages(n), cfg) }

func testAppInTempDir(t *testing.T, n int, cfg Config) *App {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return testApp(n, cfg)
}

func TestGetPages(t *testing.T) {
	t.Parallel()
	for _, n := range []int{1, 3, 6} {
		a := testApp(n, Config{})
		if got := len(a.GetPages()); got != n {
			t.Errorf("GetPages() = %d, want %d", got, n)
		}
	}
}

func TestGetPageHTML(t *testing.T) {
	t.Parallel()
	a := testApp(3, Config{})

	tests := []struct {
		index int
		empty bool
	}{
		{0, false},
		{2, false},
		{-1, true},
		{10, true},
	}
	for _, tt := range tests {
		html := a.GetPageHTML(tt.index)
		if tt.empty && html != "" {
			t.Errorf("index %d: expected empty, got %q", tt.index, html)
		}
		if !tt.empty && html == "" {
			t.Errorf("index %d: expected content, got empty", tt.index)
		}
	}
}

func TestGetFinalHTML(t *testing.T) {
	t.Parallel()
	if html := testApp(1, Config{}).GetFinalHTML(); html != "" {
		t.Errorf("no final MD: expected empty, got %q", html)
	}
	if html := testApp(1, Config{FinalMD: "# Done"}).GetFinalHTML(); html == "" {
		t.Error("with final MD: expected content, got empty")
	}
}

func TestGetHelpURL(t *testing.T) {
	t.Parallel()
	if got := testApp(1, Config{}).GetHelpURL(); got != "" {
		t.Errorf("empty config: got %q", got)
	}
	want := "https://help.example.com"
	if got := testApp(1, Config{HelpURL: want}).GetHelpURL(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOpenURLBlocked(t *testing.T) {
	a := testApp(1, Config{})
	blocked := []string{
		"javascript:alert(1)",
		"file:///etc/passwd",
		"ftp://evil.com",
		"",
		"https:opaque",
	}
	for _, u := range blocked {
		a.OpenURL(u)
	}
}

func TestOpenHelpEmpty(t *testing.T) {
	a := testApp(1, Config{})
	a.OpenHelp()
}

func TestOpenHelpBlocked(t *testing.T) {
	a := testApp(1, Config{HelpURL: "javascript:alert(1)"})
	a.OpenHelp()
}

func TestGetCheckStateEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	a := testApp(1, Config{})
	got := a.GetCheckState()
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestGetCheckStateReturnsCopy(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	a := testApp(1, Config{})
	a.ToggleCheckItem("0:0")

	m := a.GetCheckState()
	m["0:0"] = false

	if !a.GetCheckState()["0:0"] {
		t.Error("mutating returned map should not affect internal state")
	}
}

func TestToggleCheckItem(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	a := testApp(1, Config{})

	if got := a.ToggleCheckItem("0:0"); !got {
		t.Error("first toggle: want true, got false")
	}
	if got := a.ToggleCheckItem("0:0"); got {
		t.Error("second toggle: want false, got true")
	}
}

func TestToggleCheckItemPersistence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	a := testApp(1, Config{})
	a.ToggleCheckItem("0:0")

	os.Setenv("XDG_CONFIG_HOME", dir)
	a2 := testApp(1, Config{})
	if !a2.GetCheckState()["0:0"] {
		t.Error("state not persisted across app restarts")
	}
}

func TestToggleCheckItemInvalidKey(t *testing.T) {
	a := testApp(1, Config{})
	if got := a.ToggleCheckItem("invalid"); got {
		t.Error("invalid key should return false")
	}
	if got := a.ToggleCheckItem(""); got {
		t.Error("empty key should return false")
	}
	if got := a.ToggleCheckItem("abc:def"); got {
		t.Error("non-numeric key should return false")
	}
}
