# Design

Design decisions and rationale for the day1 onboarding wizard.

## Principles

1. **One page, one screen.** Pages do not scroll. Content authors must be concise. This forces clear, scannable content and prevents walls of text that new hires won't read.
2. **No framework.** The frontend is vanilla HTML/CSS/JS. No React, no build tools, no node_modules. The entire UI ships embedded in the Go binary.
3. **Runtime content.** Markdown pages are loaded from a directory at runtime (`--pages-dir`), not compiled into the binary. Content can be updated without rebuilding.
4. **Config-driven.** All content settings live in `day1.yml` alongside the pages. The CLI has only three flags: `--pages-dir`, `--force`, `--verbose`.
5. **System theme.** Light and dark themes are handled via CSS `prefers-color-scheme`, overridable in `day1.yml` with `theme: dark` or `theme: light`. On WSL, the app reads the Windows registry (`AppsUseLightTheme`) to detect dark mode since WebKit2GTK can't see the Windows theme.
6. **Embedded defaults.** Demo pages are baked into the binary via `//go:embed`. When `--pages-dir` is not set, the built-in pages are extracted to a temp dir and used automatically.

## Frontend Spec

### Window

- **Size:** 900x600, fixed, not resizable
- **Chrome:** Frameless (no OS title bar). The header area is draggable via `--wails-draggable: drag`.
- **Position:** Centered on primary display
- **Always on top:** No (unlike hermes notifications, this is a wizard the user interacts with)
- **Start hidden:** Yes. Shown after page data is loaded and window is positioned.

### Progress Bar

Horizontal stepper at the top of the window. Each page is a step:

- **Completed:** Filled green circle, green connecting line
- **Active:** Green-bordered circle with subtle glow/shadow
- **Future:** Gray circle, gray connecting line
- **Labels:** Page title below each dot (truncated with ellipsis if >90px)
- **Clickable:** Click any step to jump to that page

### Content Area

- No scrolling (overflow: hidden)
- Rendered markdown with prose-friendly typography
- Images as block elements, max height 120px
- Blockquotes styled as callout cards with green left border and tinted background
- Tables with striped rows, compact padding, uppercase muted headers
- Code blocks with monospace font, rounded corners, subtle border
- Emoji support requires `fonts-noto-color-emoji` on Linux

### Navigation

- **Next / Finish** button (bottom-right, green accent)
- **Close** button (bottom-left, muted text)
- **Enter** key advances to next page
- **Backspace** key goes back one page
- **Esc** key dismisses (closes without completing)

### Footer

The footer contains branding, help link, page indicator, and navigation buttons:

`[brand logo + name] [Need Help?] ... [3 of 6] ... [Close] [Next]`

- **Branding**: Company logo (16x16) + name, left side. Configured via `brand.name` and `brand.logo` in `day1.yml`. Hidden when not configured.
- **"Need Help?" link**: Subtle accent-colored text next to the brand. Configured via `help_url`. Opens in default browser (uses `cmd.exe /c start` on WSL). Only allows `http://`, `https://`, and `ms-settings://` schemes. Hidden when not configured.
- **Page indicator**: "N of M" centered.
- **Close button**: Hidden on the final page (only the primary button remains).

### Final Page

Built-in "You're all set!" screen with animated green checkmark circle. Overridable via `final_page` in `day1.yml` pointing to a custom markdown file.

## day1.yml

All settings live in `day1.yml` inside the pages directory:

```yaml
# day1.yml — All settings for the onboarding wizard.
brand:
  name: Example
  logo: assets/brand-logo.png # relative to this directory
title: Welcome
help_url: https://wiki.example.com/onboarding
theme: auto # auto, light, or dark
accent_color: "#188038" # hex color for buttons and progress bar
# final_page: final.md
pages: # display order; only listed pages are shown
  - welcome.md
  - getting-started.md
  - security.md
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `brand.name` | string | *(hidden)* | Company name in footer |
| `brand.logo` | string | *(hidden)* | Logo path relative to pages dir |
| `title` | string | `Welcome` | Window title |
| `help_url` | string | *(hidden)* | "Need Help?" link URL |
| `theme` | string | `auto` | `auto`, `light`, or `dark` |
| `accent_color` | string | `#188038` | Hex color for buttons and progress bar |
| `final_page` | string | *(built-in)* | Custom final page .md |
| `pages` | list | *(auto-discover)* | Ordered list of .md filenames |

**Security:** `final_page` and `pages` entries reject absolute paths and `..` traversal to prevent reading files outside the pages directory.

When `pages` is set, only listed files are loaded in that order. When omitted, all `.md` files are auto-discovered and sorted by frontmatter `order` field, then filename.

## Markdown Page Format

Each `.md` file has optional YAML frontmatter:

```yaml
---
title: Day 1         # displayed in progress bar (generated from filename if missing)
platform: all        # "all", "windows", "darwin", "linux" (default: "all")
---
```

### Content Guidelines

Since pages don't scroll, content must fit in ~400px of vertical space. Guidelines:

- One H1 heading per page
- 5-7 bullet points max, or a compact table
- One image (block, max 120px height) if desired
- Use emojis for visual interest
- Use blockquotes for callouts/tips
- Use `<span class="wave">👋</span>` for animated waving hand

## Run-Once Sentinel

- **Path:** `os.UserConfigDir()/day1/.completed` (`%AppData%\day1` on Windows, `~/Library/Application Support/day1` on macOS, `~/.config/day1` on Linux)
- **Content:** UTC timestamp in RFC 3339 format
- **Check on start:** If sentinel exists and `--force` not set, exit 0 silently
- **Write on complete:** After user clicks Close on the final page
- **Dismiss (Esc):** Does NOT write sentinel -- wizard shows again next time

## Testing Strategy

All tests are table-driven with `t.Parallel()` where safe.

| Package | What's tested | Fixtures |
|---------|---------------|----------|
| `internal/pages` | Frontmatter parsing, ordering, platform filtering, markdown rendering, image URL rewriting, title generation, config loading | `testdata/pages/`, `t.TempDir()` |
| `internal/marker` | Sentinel check/write/remove, directory creation | `t.TempDir()` |
| `internal/app` | GetPages count, GetPageHTML bounds, GetFinalHTML, GetHelpURL, URL scheme validation | In-memory test pages |
| `cmd` | Flag defaults, removed flags verification, version output, invalid pages-dir | -- |

**Coverage target:** >75% on `./internal/...`

## Patterns Provenance

| Pattern | Source |
|---------|--------|
| Wails frameless window + Cobra | hermes |
| `//go:embed all:frontend` at root | hermes |
| App struct with Wails bindings | hermes |
| Platform logging (syslog/eventlog) | hermes |
| Vanilla JS with `findApp()` discovery | hermes |
| WSL browser workaround (`cmd.exe /c start`) | hermes (adapted) |
| Version ldflags injection | converge |
| `buildRootCmd()` for testable CLI | fleet-plan |
| Table-driven tests | all three |
| CI: vet + test + codecov | hermes |
| Release: matrix wails build | hermes |

## Build

| Platform | Build tags | Notes |
|----------|-----------|-------|
| Windows | `desktop,production` | Cross-compilable from Linux (`GOOS=windows`) |
| Linux | `webkit2_41` | Requires `libgtk-3-dev`, `libwebkit2gtk-4.1-dev` |
| macOS | `desktop,production` | Must build ON macOS (CGO requires Xcode/SDK) |

## V2 Wishlist

- **Interactive elements:** Support checkboxes in markdown (`- [ ] I have read the security policy`) that gate the Next button. Ensures acknowledgment, not just page-flipping. Would require frontend state tracking and a custom goldmark extension.
