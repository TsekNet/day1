package app

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/TsekNet/day1/internal/marker"
	"github.com/TsekNet/day1/internal/pages"
	"github.com/google/deck"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type PageInfo struct {
	Title string `json:"title"`
	Index int    `json:"index"`
}

type BrandInfo struct {
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type Config struct {
	HelpURL     string
	FinalMD     string
	Theme       string
	AccentColor string
	BrandName   string
	BrandLogo   string
}

type App struct {
	ctx      context.Context
	pages    []pages.Page
	cfg      Config
	brand    BrandInfo
	rendered []string
}

func New(loaded []pages.Page, cfg Config) *App {
	rendered := make([]string, len(loaded))
	for i, p := range loaded {
		html, err := pages.RenderHTML(p.Markdown, "/pages")
		if err != nil {
			deck.Errorf("render page %s: %v", p.SourceFile, err)
			rendered[i] = "<p>Error rendering page.</p>"
			continue
		}
		rendered[i] = html
	}
	var logoURL string
	if cfg.BrandLogo != "" {
		logoURL = "/pages/" + cfg.BrandLogo
	}
	return &App{
		pages:    loaded,
		cfg:      cfg,
		brand:    BrandInfo{Name: cfg.BrandName, Logo: logoURL},
		rendered: rendered,
	}
}

func (a *App) Startup(ctx context.Context) { a.ctx = ctx }

func (a *App) GetPages() []PageInfo {
	info := make([]PageInfo, len(a.pages))
	for i, p := range a.pages {
		info[i] = PageInfo{Title: p.Frontmatter.Title, Index: i}
	}
	return info
}

func (a *App) GetPageHTML(index int) string {
	if index < 0 || index >= len(a.rendered) {
		return ""
	}
	return a.rendered[index]
}

func (a *App) GetFinalHTML() string {
	if a.cfg.FinalMD == "" {
		return ""
	}
	html, err := pages.RenderHTML(a.cfg.FinalMD, "/pages")
	if err != nil {
		deck.Errorf("render final page: %v", err)
		return ""
	}
	return html
}

func (a *App) GetHelpURL() string     { return a.cfg.HelpURL }
func (a *App) GetAccentColor() string  { return a.cfg.AccentColor }
func (a *App) GetBrand() BrandInfo     { return a.brand }

// GetTheme resolves "auto" on WSL by reading the Windows registry, since
// WebKit2GTK can't detect prefers-color-scheme from the Windows host.
func (a *App) GetTheme() string {
	if a.cfg.Theme != "auto" {
		return a.cfg.Theme
	}
	if runtime.GOOS == "linux" && isWSL() {
		if wslDarkMode() {
			return "dark"
		}
		return "light"
	}
	return "auto"
}

func (a *App) Ready() {
	wailsRuntime.WindowShow(a.ctx)
	wailsRuntime.WindowCenter(a.ctx)
}

func (a *App) Complete() {
	if err := marker.Write(); err != nil {
		deck.Errorf("write marker: %v", err)
	} else {
		deck.Info("onboarding completed, sentinel written")
	}
	wailsRuntime.Quit(a.ctx)
}

func (a *App) Dismiss() {
	deck.Info("wizard dismissed without completing")
	wailsRuntime.Quit(a.ctx)
}

func (a *App) OpenHelp() {
	if a.cfg.HelpURL == "" {
		return
	}
	u, err := url.Parse(a.cfg.HelpURL)
	if err != nil || !allowedSchemes[u.Scheme] {
		deck.Warningf("blocked URL: %s", a.cfg.HelpURL)
		return
	}
	if err := openBrowser(a.ctx, a.cfg.HelpURL); err != nil {
		deck.Errorf("open browser: %v", err)
	}
}

// openBrowser uses rundll32 on WSL to avoid cmd.exe metacharacter injection.
func openBrowser(ctx context.Context, rawURL string) error {
	if runtime.GOOS == "linux" && isWSL() {
		return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", rawURL).Start()
	}
	wailsRuntime.BrowserOpenURL(ctx, rawURL)
	return nil
}

var (
	wslOnce   sync.Once
	wslCached bool
)

func isWSL() bool {
	wslOnce.Do(func() {
		data, err := os.ReadFile("/proc/version")
		if err != nil {
			return
		}
		wslCached = strings.Contains(strings.ToLower(string(data)), "microsoft")
	})
	return wslCached
}

func wslDarkMode() bool {
	out, err := exec.Command("reg.exe", "query",
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		"/v", "AppsUseLightTheme").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "0x0")
}

var allowedSchemes = map[string]bool{"http": true, "https": true, "ms-settings": true}
