package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/TsekNet/day1/internal/app"
	"github.com/TsekNet/day1/internal/marker"
	"github.com/TsekNet/day1/internal/pages"
	"github.com/google/deck"
	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wopts "github.com/wailsapp/wails/v2/pkg/options/windows"
)

var (
	frontendAssets embed.FS
	defaultPages   embed.FS
)

var (
	flagPagesDir string
	flagForce    bool
	flagVerbose  bool
)

func Execute(assets, pages embed.FS) error {
	frontendAssets = assets
	defaultPages = pages
	return buildRootCmd().Execute()
}

// RunBindings is called during `wails build` for JS binding generation.
func RunBindings(assets embed.FS) {
	frontendAssets = assets
	stub := []pages.Page{{
		Frontmatter: pages.Frontmatter{Title: "Bindings", Order: 1},
		Markdown:    "# Generating bindings",
		SourceFile:  "stub.md",
	}}
	a := app.New(stub, app.Config{Theme: "auto"})
	wails.Run(&options.App{
		Title:       "day1",
		Width:       900,
		Height:      600,
		Frameless:   true,
		StartHidden: true,
		AssetServer: &assetserver.Options{Assets: frontendAssets},
		OnStartup:   a.Startup,
		Bind:        []interface{}{a},
		Windows:     &wopts.Options{IsZoomControlEnabled: false},
	})
}

func buildRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "day1",
		Short: "Cross-platform onboarding wizard for new hires",
		Long: `day1 renders markdown pages as a step-by-step onboarding wizard
inside a frameless webview. Write pages in markdown, configure everything
in day1.yml, and every new hire sees a polished first-run experience.

Runs with built-in demo pages by default. Use --pages-dir to point at
your own content.`,
		Example: `  day1                              # built-in demo pages
  day1 --pages-dir /opt/day1/pages
  day1 --force                      # re-show even if completed`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          run,
	}

	f := root.Flags()
	f.StringVar(&flagPagesDir, "pages-dir", "", "directory containing .md pages and day1.yml (default: built-in)")
	f.BoolVar(&flagForce, "force", false, "show even if already completed")
	f.BoolVarP(&flagVerbose, "verbose", "v", false, "verbose logging to stderr")

	root.AddCommand(versionCmd())

	return root
}

func run(cmd *cobra.Command, _ []string) error {
	if !flagForce {
		done, err := marker.Exists()
		if err != nil {
			deck.Warningf("marker check: %v", err)
		}
		if done {
			deck.Info("already completed, exiting (use --force to override)")
			return nil
		}
	}

	pagesDir := flagPagesDir
	var cleanup func()

	if pagesDir == "" {
		dir, cleanFn, err := extractEmbeddedPages()
		if err != nil {
			return fmt.Errorf("extract embedded pages: %w", err)
		}
		pagesDir = dir
		cleanup = cleanFn
		deck.Info("using built-in demo pages")
	}
	if cleanup != nil {
		defer cleanup()
	}

	cfg, err := pages.LoadConfig(pagesDir)
	if err != nil {
		deck.Warningf("config: %v (using defaults)", err)
	}

	title := firstNonEmpty(cfg.Title, "Day 1")
	theme := firstNonEmpty(cfg.Theme, "auto")

	loaded, err := pages.Load(pagesDir)
	if err != nil {
		return fmt.Errorf("load pages: %w", err)
	}
	if len(loaded) == 0 {
		return fmt.Errorf("no pages found in %s", pagesDir)
	}

	deck.Infof("loaded %d pages from %s", len(loaded), pagesDir)

	var finalMD string
	if cfg.FinalPage != "" {
		if filepath.IsAbs(cfg.FinalPage) || strings.Contains(cfg.FinalPage, "..") {
			return fmt.Errorf("final_page must be a relative path without '..'")
		}
		data, err := os.ReadFile(filepath.Join(pagesDir, cfg.FinalPage))
		if err != nil {
			return fmt.Errorf("read final page: %w", err)
		}
		finalMD = string(data)
	}

	a := app.New(loaded, app.Config{
		HelpURL:     cfg.HelpURL,
		FinalMD:     finalMD,
		Theme:       theme,
		AccentColor: cfg.AccentColor,
		BrandName:   cfg.Brand.Name,
		BrandLogo:   cfg.Brand.Logo,
	})

	if runtime.GOOS == "linux" {
		if os.Getenv("XDG_SESSION_TYPE") == "wayland" || os.Getenv("WAYLAND_DISPLAY") != "" {
			deck.Info("wayland session detected, forcing GDK_BACKEND=x11 for window positioning")
			os.Setenv("GDK_BACKEND", "x11")
		}
	}

	pagesHandler := http.StripPrefix("/pages/", http.FileServer(http.Dir(pagesDir)))

	err = wails.Run(&options.App{
		Title:         title,
		Width:         900,
		Height:        600,
		Frameless:     true,
		DisableResize: true,
		StartHidden:   true,
		AssetServer: &assetserver.Options{
			Assets:  frontendAssets,
			Handler: pagesHandler,
		},
		OnStartup: a.Startup,
		Bind:      []interface{}{a},
		Windows:   &wopts.Options{IsZoomControlEnabled: false},
	})
	if err != nil {
		return fmt.Errorf("wails: %w", err)
	}
	return nil
}

// extractEmbeddedPages extracts to a temp dir. Caller must call cleanup.
func extractEmbeddedPages() (string, func(), error) {
	tmp, err := os.MkdirTemp("", "day1-pages-*")
	if err != nil {
		return "", nil, err
	}

	err = fs.WalkDir(defaultPages, "testdata/pages", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("testdata/pages", path)
		target := filepath.Join(tmp, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		data, err := defaultPages.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
	if err != nil {
		os.RemoveAll(tmp)
		return "", nil, err
	}

	return tmp, func() { os.RemoveAll(tmp) }, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
