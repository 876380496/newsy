# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- Install/update dependencies: `go mod tidy`
- Run the app: `go run ./cmd/newsy`
- Run all tests: `go test ./...`
- Run one package's tests: `go test ./internal/source/...`
- Run one specific test: `go test ./internal/source -run TestValidateSpecs`
- Build the binary: `go build ./cmd/newsy`

## Runtime notes

- The app now uses XDG-style user directories by default instead of the repository root.
- Default config path: `~/.config/newsy/config.yaml`.
- Default plugin directory: `~/.config/newsy/plugins`.
- Default SQLite path: `~/.local/share/newsy/newsy.db`.
- Default log path: `~/.cache/newsy/newsy.log`.
- You can inspect the active paths with `go run ./cmd/newsy --print-paths`.
- This is a Bubble Tea TUI app, so `go run` needs a real interactive terminal; non-TTY environments can fail during startup.

## Architecture overview

This project is a modular TUI news reader written in Go. RSS is the current built-in source, but the core design is provider-based so new source types can be added without rewriting the app layer.

### Application wiring

- `cmd/newsy/main.go` is a thin entrypoint that delegates startup to `internal/app`.
- `internal/app/app.go` wires startup: load config, build source specs, open SQLite, register providers, validate specs, perform initial refresh, create the UI model, and run Bubble Tea.
- `internal/app/actions.go` is the orchestration layer after startup. The UI emits typed actions; the app handles refresh, filtering, read/star toggles, browser open, and article reloads.

### Source/provider architecture

The source system is intentionally not RSS-specific.

- `internal/source/spec.go` defines runtime source definitions from config.
- `internal/source/provider.go` defines the provider contract: `Type()`, `Validate(spec)`, and `Fetch(ctx, spec)`.
- `internal/source/registry.go` stores providers by type.
- `internal/source/service.go` is the boundary that validates config, fetches articles, persists them, and records fetch state.
- `internal/source/validate.go` validates all configured specs against the registry at startup.
- `internal/providers/register.go` registers built-in providers.
- `internal/plugin/plugin.go` is the future plugin boundary: plugins contribute providers through the same registry path.
- `internal/source/rss` implements the current RSS provider via `github.com/mmcdole/gofeed` and maps feed items into `domain.Article`.

### Data model and persistence

- `internal/domain/article.go` defines the shared article model used across providers, storage, and UI.
- `internal/store/sqlite.go` opens SQLite and applies schema creation.
- `internal/store/articles.go` owns article upsert/list/read/star persistence.
- `internal/store/state.go` tracks per-source fetch success/error state.

Important boundary: source definitions live in `config.yaml`; SQLite stores fetched content and fetch state only.

### UI structure

- `internal/ui` contains the Bubble Tea model, messages, keymap, update loop, and rendering.
- The UI package is stateful but intentionally dumb about business logic: it emits actions and renders data; `internal/app` owns behavior.
- The layout is a three-pane interface: sources, articles, preview.
- Pane rendering in `internal/ui/view.go` depends on a layered lipgloss pattern: inner content styles handle width/height/max-height clipping, while outer styles only apply borders/padding. Keep that split when editing layout code.

### Filtering and article flow

The article list shown in the UI is composed in `internal/app/actions.go` by applying filters in order:

1. selected source
2. unread/starred flags
3. title search

`App.allArticles` stores the full in-memory article set loaded from SQLite, and the filtered subset is pushed into the UI model.

### Content preview behavior

- The preview pane prefers `Article.Content`, falls back to `Article.Summary`, and strips HTML before rendering.
- RSS feeds can contain large HTML fragments, so preview cleanup happens before display instead of rendering raw feed markup directly.

## Working conventions for this repo

- Keep the code modular and simple; avoid coupling RSS-specific behavior into core app/source abstractions.
- When adding a new source type, implement a new `source.Provider` and register it through the existing registry path rather than branching inside the app layer.
- When changing TUI pane layout, preserve the inner-content-clipping / outer-border-rendering split in `internal/ui/view.go`.
