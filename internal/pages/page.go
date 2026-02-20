// Package pages discovers, parses, and renders markdown pages for the
// day1 onboarding wizard.
package pages

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	Title    string `yaml:"title"`
	Order    int    `yaml:"order"`
	Platform string `yaml:"platform"`
}

type Page struct {
	Frontmatter Frontmatter
	Markdown    string
	SourceFile  string
}

var (
	fmDelim  = regexp.MustCompile(`(?m)^---\s*$`)
	imgSrcRe = regexp.MustCompile(`(<img\s[^>]*?src=")([^"]+)(")`)
	renderer = goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Typographer))
)

// ParseFrontmatter splits raw markdown into YAML frontmatter + body.
// Returns platform="all" if no delimiters are found.
func ParseFrontmatter(raw, filename string) (Frontmatter, string, error) {
	locs := fmDelim.FindAllStringIndex(raw, 3)
	if len(locs) < 2 {
		return Frontmatter{Platform: "all"}, raw, nil
	}

	fmBlock := raw[locs[0][1]:locs[1][0]]
	body := raw[locs[1][1]:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(fmBlock), &fm); err != nil {
		return Frontmatter{}, "", fmt.Errorf("parse frontmatter in %s: %w", filename, err)
	}
	if fm.Platform == "" {
		fm.Platform = "all"
	}
	return fm, body, nil
}

// RenderHTML converts markdown to HTML. assetsPrefix is prepended to relative
// image src attributes so the Wails AssetHandler can serve them.
func RenderHTML(markdown, assetsPrefix string) (string, error) {
	var buf bytes.Buffer
	if err := renderer.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("goldmark: %w", err)
	}
	if assetsPrefix == "" {
		return buf.String(), nil
	}
	return rewriteImageSrcs(buf.String(), assetsPrefix), nil
}

var skipPrefixes = []string{"http://", "https://", "//", "/", "data:"}

func rewriteImageSrcs(html, prefix string) string {
	prefix = strings.TrimSuffix(prefix, "/")
	return imgSrcRe.ReplaceAllStringFunc(html, func(match string) string {
		g := imgSrcRe.FindStringSubmatch(match)
		if len(g) != 4 {
			return match
		}
		for _, s := range skipPrefixes {
			if strings.HasPrefix(g[2], s) {
				return match
			}
		}
		return g[1] + prefix + "/" + g[2] + g[3]
	})
}
