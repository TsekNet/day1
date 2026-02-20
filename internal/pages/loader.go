package pages

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func Load(dir string) ([]Page, error) {
	return LoadForPlatform(dir, runtime.GOOS)
}

func LoadForPlatform(dir, platform string) ([]Page, error) {
	cfg, _ := LoadConfig(dir)
	if len(cfg.Pages) > 0 {
		return loadList(dir, cfg.Pages, platform)
	}
	return loadAll(dir, platform)
}

func loadList(dir string, names []string, platform string) ([]Page, error) {
	out := make([]Page, 0, len(names))
	for _, name := range names {
		if strings.Contains(name, "..") || filepath.IsAbs(name) {
			return nil, fmt.Errorf("invalid page path: %s", name)
		}
		p, err := readPage(dir, name, platform)
		if err != nil {
			return nil, err
		}
		if p != nil {
			out = append(out, *p)
		}
	}
	return out, nil
}

// loadAll is the fallback when day1.yml has no `pages:` list.
func loadAll(dir, platform string) ([]Page, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read pages dir %s: %w", dir, err)
	}

	var out []Page
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p, err := readPage(dir, e.Name(), platform)
		if err != nil {
			return nil, err
		}
		if p != nil {
			out = append(out, *p)
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Frontmatter.Order != out[j].Frontmatter.Order {
			return out[i].Frontmatter.Order < out[j].Frontmatter.Order
		}
		return out[i].SourceFile < out[j].SourceFile
	})
	return out, nil
}

// readPage returns nil (not error) when filtered out by platform.
func readPage(dir, name, platform string) (*Page, error) {
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}

	fm, body, err := ParseFrontmatter(string(raw), name)
	if err != nil {
		return nil, err
	}

	if fm.Platform != "all" && fm.Platform != platform {
		return nil, nil
	}

	if fm.Title == "" {
		fm.Title = titleFromFilename(name)
	}

	return &Page{Frontmatter: fm, Markdown: body, SourceFile: name}, nil
}

// titleFromFilename: "tools-access.md" -> "Tools Access"
func titleFromFilename(name string) string {
	name = strings.TrimSuffix(name, ".md")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	words := strings.Fields(name)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
