package pages

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		wantTitle string
		wantOrder int
		wantPlat  string
		wantBody  string
		wantErr   bool
	}{
		{
			name: "full frontmatter",
			raw: "---\ntitle: Welcome\norder: 1\nplatform: darwin\n---\n# Hello\n",
			wantTitle: "Welcome",
			wantOrder: 1,
			wantPlat:  "darwin",
			wantBody:  "# Hello\n",
		},
		{
			name:      "no frontmatter",
			raw:       "# Just markdown\nSome text.",
			wantTitle: "",
			wantOrder: 0,
			wantPlat:  "all",
			wantBody:  "# Just markdown\nSome text.",
		},
		{
			name: "frontmatter without platform defaults to all",
			raw:  "---\ntitle: Setup\norder: 2\n---\nContent here.",
			wantTitle: "Setup",
			wantOrder: 2,
			wantPlat:  "all",
			wantBody:  "Content here.",
		},
		{
			name:    "invalid yaml",
			raw:     "---\n: [broken\n---\nBody",
			wantErr: true,
		},
		{
			name:      "empty file",
			raw:       "",
			wantTitle: "",
			wantPlat:  "all",
			wantBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm, body, err := ParseFrontmatter(tt.raw, "test.md")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fm.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", fm.Title, tt.wantTitle)
			}
			if fm.Order != tt.wantOrder {
				t.Errorf("order = %d, want %d", fm.Order, tt.wantOrder)
			}
			if fm.Platform != tt.wantPlat {
				t.Errorf("platform = %q, want %q", fm.Platform, tt.wantPlat)
			}
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestRenderHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		markdown     string
		assetsPrefix string
		wantContains []string
	}{
		{
			name:         "heading",
			markdown:     "# Hello World",
			wantContains: []string{"<h1>Hello World</h1>"},
		},
		{
			name:         "bullet list",
			markdown:     "- one\n- two\n- three",
			wantContains: []string{"<ul>", "<li>one</li>", "<li>three</li>"},
		},
		{
			name:         "image without prefix",
			markdown:     "![logo](assets/logo.png)",
			wantContains: []string{`src="assets/logo.png"`},
		},
		{
			name:         "image with prefix",
			markdown:     "![logo](assets/logo.png)",
			assetsPrefix: "/pages",
			wantContains: []string{`src="/pages/assets/logo.png"`},
		},
		{
			name:         "absolute image unchanged",
			markdown:     "![pic](https://example.com/img.png)",
			assetsPrefix: "/pages",
			wantContains: []string{`src="https://example.com/img.png"`},
		},
		{
			name:         "gfm table",
			markdown:     "| A | B |\n|---|---|\n| 1 | 2 |",
			wantContains: []string{"<table>", "<td>1</td>"},
		},
		{
			name:         "bold and italic",
			markdown:     "**bold** and *italic*",
			wantContains: []string{"<strong>bold</strong>", "<em>italic</em>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := RenderHTML(tt.markdown, tt.assetsPrefix)
			if err != nil {
				t.Fatalf("RenderHTML: %v", err)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\ngot: %s", want, got)
				}
			}
		})
	}
}

func TestLoadForPlatform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		files      map[string]string
		platform   string
		wantCount  int
		wantTitles []string
	}{
		{
			name: "all pages on matching platform",
			files: map[string]string{
				"01-first.md":  "---\ntitle: First\norder: 1\n---\nContent",
				"02-second.md": "---\ntitle: Second\norder: 2\n---\nMore",
			},
			platform:   "linux",
			wantCount:  2,
			wantTitles: []string{"First", "Second"},
		},
		{
			name: "platform filtering",
			files: map[string]string{
				"01-mac.md":  "---\ntitle: Mac Only\norder: 1\nplatform: darwin\n---\n",
				"02-all.md":  "---\ntitle: Everyone\norder: 2\n---\n",
				"03-win.md":  "---\ntitle: Win Only\norder: 3\nplatform: windows\n---\n",
			},
			platform:   "darwin",
			wantCount:  2,
			wantTitles: []string{"Mac Only", "Everyone"},
		},
		{
			name: "order from frontmatter beats filename",
			files: map[string]string{
				"z-last.md":  "---\ntitle: Actually First\norder: 1\n---\n",
				"a-first.md": "---\ntitle: Actually Second\norder: 2\n---\n",
			},
			platform:   "linux",
			wantCount:  2,
			wantTitles: []string{"Actually First", "Actually Second"},
		},
		{
			name: "title generated from filename when missing",
			files: map[string]string{
				"tools-access.md": "---\norder: 1\n---\nContent",
			},
			platform:   "linux",
			wantCount:  1,
			wantTitles: []string{"Tools Access"},
		},
		{
			name: "non-md files ignored",
			files: map[string]string{
				"page.md":    "---\ntitle: Page\norder: 1\n---\n",
				"readme.txt": "not a page",
				"logo.png":   "binary",
			},
			platform:   "linux",
			wantCount:  1,
			wantTitles: []string{"Page"},
		},
		{
			name:      "empty directory",
			files:     map[string]string{},
			platform:  "linux",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			got, err := LoadForPlatform(dir, tt.platform)
			if err != nil {
				t.Fatalf("LoadForPlatform: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Fatalf("got %d pages, want %d", len(got), tt.wantCount)
			}
			for i, want := range tt.wantTitles {
				if got[i].Frontmatter.Title != want {
					t.Errorf("page[%d].Title = %q, want %q", i, got[i].Frontmatter.Title, want)
				}
			}
		})
	}
}

func TestLoadTestdata(t *testing.T) {
	t.Parallel()

	root := testdataRoot(t)
	pagesDir := filepath.Join(root, "testdata", "pages")

	pages, err := LoadForPlatform(pagesDir, runtime.GOOS)
	if err != nil {
		t.Fatalf("LoadForPlatform(testdata): %v", err)
	}
	if len(pages) == 0 {
		t.Fatal("expected at least one page in testdata/pages")
	}

	for _, p := range pages {
		if p.Frontmatter.Title == "" {
			t.Errorf("page %s has empty title", p.SourceFile)
		}
		html, err := RenderHTML(p.Markdown, "/pages")
		if err != nil {
			t.Errorf("RenderHTML(%s): %v", p.SourceFile, err)
		}
		if html == "" {
			t.Errorf("page %s rendered to empty HTML", p.SourceFile)
		}
	}
}

func TestTitleFromFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename string
		want     string
	}{
		{"welcome.md", "Welcome"},
		{"tools-access.md", "Tools Access"},
		{"getting-started.md", "Getting Started"},
		{"simple.md", "Simple"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			got := titleFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("titleFromFilename(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		yaml      string
		wantBrand string
		wantHelp  string
		wantTheme string
		wantErr   bool
	}{
		{
			name:      "full config",
			yaml:      "brand:\n  name: Acme\n  logo: img/logo.png\nhelp_url: https://help.acme.com\ntheme: dark\n",
			wantBrand: "Acme",
			wantHelp:  "https://help.acme.com",
			wantTheme: "dark",
		},
		{
			name:      "brand only",
			yaml:      "brand:\n  name: TestCorp\n",
			wantBrand: "TestCorp",
		},
		{
			name: "empty file",
			yaml: "",
		},
		{
			name:    "invalid yaml",
			yaml:    ": [broken",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if tt.yaml != "" || tt.wantErr {
				os.WriteFile(dir+"/day1.yml", []byte(tt.yaml), 0o644)
			}
			cfg, err := LoadConfig(dir)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Brand.Name != tt.wantBrand {
				t.Errorf("brand.name = %q, want %q", cfg.Brand.Name, tt.wantBrand)
			}
			if cfg.HelpURL != tt.wantHelp {
				t.Errorf("help_url = %q, want %q", cfg.HelpURL, tt.wantHelp)
			}
			if cfg.Theme != tt.wantTheme {
				t.Errorf("theme = %q, want %q", cfg.Theme, tt.wantTheme)
			}
		})
	}
}

func TestLoadConfigMissing(t *testing.T) {
	t.Parallel()
	cfg, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatalf("missing config should not error: %v", err)
	}
	if cfg.Brand.Name != "" {
		t.Errorf("expected empty brand name, got %q", cfg.Brand.Name)
	}
}

// testdataRoot walks up from the test file to the repo root.
func testdataRoot(t *testing.T) string {
	t.Helper()
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod)")
		}
		dir = parent
	}
}
