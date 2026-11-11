# Agent Orientation

Applies to every task in this repo. Detailed rules: `.clinerules/01`–`08`. Fork invariants are machine-checked: `fork-check .`.

## What this is

A hard fork of [sentriz/gonic](https://github.com/sentriz/gonic) — a free-software Subsonic server in Go — with extra admin UX and features upstream has been slow to take. The goal is to keep those fork features while selectively syncing upstream fixes.

## Removed

- Nothing from upstream's feature set. This fork is additive; it deletes no Subsonic endpoint.
- No client app or JS toolchain: the admin UI is server-side Go templates.
- No upstream badge block in `README.md` (no Docker Hub / repology badges).

## Added

- Public tokenized share links: landing page, in-browser streaming, downloads, and the full Subsonic share API.
- Smart playlists: m3u files carrying a `#GONIC-RULES` attribute are expanded into track queries.
- Admin pages: shares, playlists (with rules editor), podcasts, internet radio, live scan status, scan schedule, transcode cache size/ejection.
- Dark-only admin theme.
- OPML podcast import.

## Kept

- Every Subsonic endpoint, the token+salt auth model, folder and tag browsing, transcoding, cover cache, jukebox, podcasts, last.fm and ListenBrainz scrobbling.
- Upstream's gorm v1 + SQLite + gormigrate persistence.

## Stack

- Go 1.27, no web framework (`net/http` + `http.ServeMux`), gorm v1, SQLite (CGO-free ncruces driver by default).
- Entry point `cmd/gonic/`; Subsonic API `server/ctrlsubsonic/`; admin `server/ctrladmin/`; share links `server/ctrlshare/`.
- Persistence and migrations `db/`; admin templates + prebuilt tailwind `server/ctrladmin/adminui/`.

## Gates

```bash
CGO_ENABLED=0 go build ./...
go vet ./...
go test ./...
golangci-lint run --config .golangci.yml   # CI pins v2.14.0
fork-check .                               # fork invariants
```

No push, no tag, until all of these pass. See `.clinerules/03-workflow.md`.

## Sync

Upstream is `sentriz/gonic`. Fork commits are **cherry-picked**, never merged, and the GitHub "Sync fork" button is never used. Candidates come from `git cherry -v upstream/master HEAD`, not raw rev-ranges. Full procedure and protected features: `.clinerules/08-upstream-sync.md`.

## Non-negotiables

- Never push to upstream's registry namespace; this fork publishes to `ghcr.io/antebudimir/gonic` only (`.clinerules/06-ci.md`).
- Never add `merge=ours` attributes — cherry-pick uses per-file merge machinery, so `ours` would silently discard upstream changes.
- Never let a test touch a real database file; use `db.NewMock(":memory:", ...)` or `t.TempDir()`.
- Never tag a commit that has not passed the gates.
- Never delete a fork feature to satisfy a request; propose a fork-shaped alternative.

## Audience split

- `README.md` fork section is for people: user-facing behaviour only, no process meta-language, no links to this file.
- `AGENTS.md` and `.clinerules/` are for agents: operational knowledge, kept short and imperative.
