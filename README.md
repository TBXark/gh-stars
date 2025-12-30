# gh-stars

A Fyne MVVM best-practice example using a GitHub Stars viewer as the sample app.

## What this project demonstrates

- MVVM layering with Fyne data binding
- Clear service/client boundaries (UI decoupled from HTTP)
- A Navigator for window/page routing and repo details
- Async loading with cancellation

## Run

```bash
go run ./cmd/stars-gui
```

Token is optional, but recommended to increase GitHub API rate limits.

## Structure

- `cmd/stars-gui/main.go`: entry point and wiring
- `internal/app`: use cases (stars / repos)
- `internal/github`: GitHub API client
- `internal/ui`: Fyne View / ViewModel / Navigator

## Docs

- `docs/Architecture.md`
- `docs/Navigator.md`
