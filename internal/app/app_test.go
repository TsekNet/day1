package app

import (
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

func TestAllowedSchemes(t *testing.T) {
	t.Parallel()
	for scheme, want := range map[string]bool{
		"http": true, "https": true, "ms-settings": true,
		"ftp": false, "file": false, "javascript": false, "": false,
	} {
		if got := allowedSchemes[scheme]; got != want {
			t.Errorf("allowedSchemes[%q] = %v, want %v", scheme, got, want)
		}
	}
}
