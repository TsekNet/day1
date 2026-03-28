# Day1

Cross-platform onboarding wizard: markdown pages in a frameless Wails v2 webview. Runs once per machine, then exits.

## Architecture

Read `docs/architecture.md` (startup flow), `docs/design.md` (frontend spec, platform behavior), `docs/deployment.md` (run-once triggers per OS).

| Package | Purpose |
| --- | --- |
| `internal/app` | Page state, checkbox persistence, theming, WSL dark mode, browser opens |
| `internal/pages` | Markdown discovery, YAML frontmatter, goldmark rendering |
| `internal/marker` | Run-once sentinel (`.completed` in user config dir) |
| `internal/urischeme` | URI whitelist for browser opens |
| `internal/logging` | Platform-specific: syslog (unix), Windows Event Log |
| `frontend/` | Vanilla HTML/CSS/JS, embedded in binary, no build tools |

## Build and test

```bash
go vet -tags webkit2_41 ./...
go test -v -race -count=1 -tags webkit2_41 ./internal/...
wails build -skipbindings -platform linux/amd64 -tags webkit2_41 \
  -ldflags "-s -w -X github.com/TsekNet/day1/internal/version.Version=dev"
```

Linux CI needs `libgtk-3-dev libwebkit2gtk-4.1-dev`. `-tags webkit2_41` is Linux-only. Windows builds need `-windowsconsole`.

## Agents

Three agents in `.claude/agents/`. Use when changes touch their domain:

| Agent | Trigger files |
| --- | --- |
| `platform-reviewer` | `internal/logging/`, `internal/app/` (WSL/browser), `runtime.GOOS` branches, `build/` |
| `content-reviewer` | `internal/pages/`, `testdata/pages/`, `day1.yml`, frontmatter, markdown rendering |
| `integration-tester` | After code changes: build, run with demo pages, validate startup and marker behavior |
