# CLAUDE.md — ScreenShare App

## Project Overview

A self-contained WebRTC screen-sharing application. A Go backend serves as a signaling server (WebSocket) and static file host. Broadcasters share their screen; viewers connect and watch via peer-to-peer WebRTC.

## Tech Stack

- **Backend:** Go 1.22, single `main.go`, gorilla/websocket
- **Frontend:** Vanilla HTML/CSS/JS (no framework, no bundler)
- **Build:** Makefile, cross-compiled to Windows/Linux/macOS
- **Embedding:** Static files in `public/` are embedded into the binary via `//go:embed`

## Project Structure

```
main.go              — Go signaling server (WebSocket + HTTP)
config.json          — Runtime config (port, STUN server)
Makefile             — Build targets (build, build-all, clean, run, icon)
public/
  index.html         — Landing page with links to broadcaster/viewer
  broadcaster.html   — Broadcaster page (screen capture + WebRTC)
  viewer.html        — Viewer page (connect to broadcaster + WebRTC)
  common.js          — Shared utilities (status, WebSocket URL, cast, config)
  style.css          — Global styles
  favicon.ico        — Favicon
  app-icon.png       — App icon
winres/              — Windows resource files (icon embedding)
dist/                — Build output (cross-platform binaries)
```

## Key Commands

```sh
make run             # Run the server locally (go run .)
make build           # Build for current platform
make build-all       # Cross-compile for windows/linux/darwin (amd64 + arm64)
make clean           # Remove dist/
make icon            # Generate Windows icon resources
```

## Architecture Notes

- The Go server handles both HTTP file serving and WebSocket signaling on the same port.
- WebSocket connects to the root path (`/`); the handler checks for upgrade headers.
- Clean URL routes: `/broadcaster` and `/viewer` rewrite to their `.html` files.
- Client IDs are random 16-char hex strings.
- The signaling protocol is simple JSON: register, get-broadcasters, offer, answer, ice-candidate.
- STUN server is configurable via `config.json` or `--stun` flag, served to clients via `/config`.
- Config priority: defaults -> config.json -> CLI flags.

## Conventions

- No external JS dependencies — keep it vanilla.
- All frontend HTML files load `common.js` for shared utilities.
- CSS variables defined in `:root` in `style.css` for theming.
- Go code uses a global `sync.RWMutex` for the client/broadcaster maps, with per-client mutexes for WebSocket writes.
- Embed all static assets into the Go binary for single-file distribution.

## Common Pitfalls

- After editing files in `public/`, you must rebuild the binary for changes to take effect in the embedded FS. During development, use `make run` (go run) which re-embeds on each run.
- The `stunServer` value in config.json must use the `stun:` URI scheme (e.g., `stun:stun.l.google.com:19302`).
- WebSocket messages use `json.RawMessage` for offer/answer/candidate fields to relay them without re-serialization.
