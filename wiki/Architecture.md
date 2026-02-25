# Architecture

**[← Wiki Home](Home)** · [Design](Design) · [Deployment](Deployment)

---

## Startup Flow

```mermaid
flowchart TD
    Start["main()"] --> Logging["Init deck logging\nsyslog / eventlog"]
    Logging --> Cobra["Cobra CLI\nparse --pages-dir, --force, --verbose"]
    Cobra --> SentinelCheck{"Sentinel\nexists?"}
    SentinelCheck -->|"yes + no --force"| SilentExit["Exit 0"]
    SentinelCheck -->|"no or --force"| LoadConfig["Load day1.yml\nbrand, theme, help_url, pages"]
    LoadConfig --> LoadPages["Load .md files\nin day1.yml order"]
    LoadPages --> ParseFM["Parse YAML frontmatter\nfilter by platform"]
    ParseFM --> RenderMD["Render markdown\nvia goldmark"]
    RenderMD --> WailsRun["wails.Run()\n900x600 frameless"]
    WailsRun --> ShowWindow["Center + show window"]
```

---

## Package Dependency Graph

```mermaid
flowchart LR
    main["main.go"] --> cmd["cmd/"]
    cmd --> app["internal/app"]
    cmd --> marker["internal/marker"]
    cmd --> pagesP["internal/pages"]
    app --> pagesP
    app --> marker
    main --> logging["internal/logging"]
    cmd --> version["internal/version"]
```

---

## Page Lifecycle

```mermaid
flowchart LR
    Config["day1.yml\npages list"] --> Ordered["Load in order"]
    Ordered --> MDFile[".md file\non disk"]
    MDFile --> Frontmatter["Parse YAML\nfrontmatter"]
    Frontmatter --> Filter["Filter by\nruntime.GOOS"]
    Filter --> Goldmark["Render HTML\nvia goldmark"]
    Goldmark --> RewriteURLs["Rewrite image\nsrc paths"]
    RewriteURLs --> Cache["Cache in\nApp.rendered"]
    Cache --> Binding["GetPageHTML(i)\nWails binding"]
    Binding --> InnerHTML["Frontend\ninnerHTML"]
```

---

## Frontend State Machine

```mermaid
stateDiagram-v2
    [*] --> Init: DOMContentLoaded
    Init --> PageView: GetPages + GetPageHTML(0)
    PageView --> PageView: Next/Enter (i < total-1)
    PageView --> PageView: Backspace (i > 0)
    PageView --> PageView: Click step dot
    PageView --> FinalPage: Next/Enter (i == total-1)
    FinalPage --> Completed: Next/Close
    PageView --> Dismissed: Esc/Close
    Completed --> [*]: Write sentinel + quit
    Dismissed --> [*]: Quit (no sentinel)
```

---

## Build and Release Pipeline

```mermaid
flowchart LR
    Push["Push v* tag"] --> Matrix["Matrix CI\nwin / mac / linux"]
    Matrix --> WailsBuild["wails build\n-ldflags version"]
    WailsBuild --> Artifacts["Upload\nplatform binaries"]
    Artifacts --> Release["GitHub Release\n+ checksums"]
```

---

## Key Files

| File | Responsibility |
|------|----------------|
| `main.go` | Embeds frontend + demo pages, inits logging, calls `cmd.Execute()` |
| `cmd/root.go` | Cobra root command, loads `day1.yml`, launches Wails |
| `cmd/version.go` | Version subcommand |
| `internal/app/app.go` | Wails App struct, JS bindings, sentinel write on complete, WSL browser workaround |
| `internal/pages/config.go` | Parse `day1.yml` (brand, theme, accent_color, help_url, pages order, final_page) |
| `internal/pages/loader.go` | Load `.md` files in `day1.yml` order or auto-discover, platform filtering |
| `internal/pages/page.go` | Frontmatter parsing, goldmark rendering, image URL rewriting |
| `internal/marker/marker.go` | Sentinel file check/write/remove |
| `internal/logging/unix.go` | Syslog backend for macOS/Linux |
| `internal/logging/windows.go` | Event Log backend for Windows |
| `internal/version/version.go` | Version/Commit/Date vars (ldflags) |
| `frontend/index.html` | HTML structure: brand, progress bar, content area, nav |
| `frontend/style.css` | Light/dark theme, markdown typography, no-scroll design |
| `frontend/main.js` | Wails bindings, navigation, keyboard handlers, theme application |
