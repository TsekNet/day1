package app

import (
	"testing"

	"github.com/TsekNet/day1/internal/pages"
	"github.com/TsekNet/day1/internal/urischeme"
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

func TestURISchemeAllowed(t *testing.T) {
	t.Parallel()
	for url, want := range map[string]bool{
		"https://example.com":          true,
		"http://intranet/kb":           true,
		"ms-settings:windowsupdate":    urischeme.AllowedOn("ms-settings:windowsupdate", "windows"),
		"ftp://evil.com":               false,
		"file:///etc/passwd":           false,
		"javascript:alert(1)":          false,
		"":                             false,
	} {
		if got := urischeme.AllowedOn(url, "linux"); got != want {
			t.Errorf("AllowedOn(%q, linux) = %v, want %v", url, got, want)
		}
	}
}

