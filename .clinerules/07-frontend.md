# Admin UI (Go templates + tailwind)

## Stack

- Go `html/template`, embedded via `//go:embed` in `server/ctrladmin/adminui/adminui.go`.
- Tailwind v3.2.4, prebuilt to `static/style.css`; source is `style.css` + `tailwind.config.js` in the same directory.
- No JS framework, no package manager, no i18n catalogue: copy is English literals inside the templates.

## Gate

- `CGO_ENABLED=0 go build ./... && go test ./...` before commit.
- Templates are parsed at runtime, so a compile pass is not proof: load the page you changed against a scratch DB.
- After CSS changes run `go generate ./server/ctrladmin/adminui/` (it shells out to `npx tailwindcss@v3.2.4`) and commit the regenerated `static/style.css` together with the source change.

## Contracts

- Page data is `templateData` (`server/ctrladmin/ctrl.go`). A new page means a new field there, a handler in `handlers*.go`, and a route in `ctrl.go`.
- Partial state is JSON: scan status is polled from `/scan_status`; keep the response shape and the template script in sync.
- Optional surfaces follow their flags: `-scan-non-admin-enabled` (non-admin `startScan`), `-jukebox-enabled` (jukebox). Podcasts (`-podcast-path`), cache (`-cache-path`) and playlists (`-playlists-path`) are mandatory — the process exits if they are empty.

## Conventions

- Copy the markup of the nearest existing page; do not invent new component classes.
- Forms post to `*_do` routes and use the existing `userChain` / `adminChain` middleware in `ctrl.go`; do not add a second auth path.
- Keep destructive actions on `adminChain`.
