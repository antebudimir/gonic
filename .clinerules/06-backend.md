# Backend (Go)

## Stack

- Go 1.27 (`go.mod`); no web framework — `net/http` handlers and `http.ServeMux` registration.
- gorm v1 (`github.com/jinzhu/gorm`), SQLite. CGO-free by default via `github.com/ncruces/go-sqlite3`; `mattn/go-sqlite3` behind the `nowasm` build tag (`deps/`).
- Migrations: `gopkg.in/gormigrate.v1`, in `db/migrations.go`.
- Linter: golangci-lint v2.14.0, config `.golangci.yml`.

## Migrations

- One migration = one `construct(ctx, "<YYYYMMDDHHMM>", migrateX)` entry appended to the list in `Migrate` (`db/migrations.go`); use `constructNoTx` only when a step cannot run in a transaction.
- IDs are timestamps, monotonically increasing; get the current one with `date '+%Y%m%d%H%M'`.
- New model: `tx.AutoMigrate(MyModel{})` inside the migration func.
- Changed struct: rename or backfill columns **before** `AutoMigrate`, or gorm creates the new column empty and the rename collides (see `migrateArtistInfoFieldsBySource`).
- There are no down migrations and `UseTransaction: false`: get it right the first time.
- Verify: `go test ./db/...`, then start the binary against a **copy** of a real `gonic.db` and confirm it boots and serves `/rest/ping`.

## Testing

- Gate: `go test ./...` before any commit.
- Test isolation: never point a test at a real database. Use `db.NewMock(":memory:", ...)` (`db/db.go`) or a `t.TempDir()` path.
- Fixtures for Subsonic responses live in `server/ctrlsubsonic/testdata/`; regenerate deliberately, never by hand-editing expected JSON to match broken output.
- Parallel tests: `t.Parallel()` when there is no shared state.

## Verification before tag or push

```bash
CGO_ENABLED=0 go build ./...
go vet ./...
go test ./...
golangci-lint run --config .golangci.yml
fork-check .
```

## Auth and secrets

- Subsonic token+salt auth (`t`/`s`, legacy `p`); sessions via `github.com/gorilla/sessions` + `gormstore`.
- No repo-managed pepper or signing secret: session keys are per-deployment configuration, not committed.
- Credentials in query strings (`p`, `t`, `s`, `apiKey`) are redacted by `handlerutil.Log`; keep it that way (`fork.conf`).

## CI and registry

See `06-ci.md` — it is scoped to `.github/workflows/**` and is the single authoritative source.

## Dependency management

- Modules: `go.mod` / `go.sum`. Update with `go get -u ./...`, tidy with `go mod tidy`.
- A dependency swap of the SQL driver or lockfile is high risk; smoke-test against a real DB before shipping (`08-upstream-sync.md`).

## Code map

- Handlers: `server/ctrlsubsonic/handlers_*.go` (Subsonic), `server/ctrladmin/handlers*.go` (admin), `server/ctrlshare/ctrlshare.go` (public share links).
- Params: `server/ctrlsubsonic/params/params.go`.
- Models: `db/` (`Artist`, `Album`, `Track`, `TrackPlay`, `Share`, `ShareEntry`, `Podcast`, …).
- Routes: `Handle("METHOD /path", chain)` in `server/ctrlsubsonic/ctrl.go` and `server/ctrladmin/ctrl.go`.

## Fork routes to re-check after every upstream sync

`/shares`, `/playlists`, `/update_playlist_rules_do`, `/scan_status`, `/update_scan_schedule_do`, `/eject_transcode_cache_do`, `/update_transcode_cache_size_do`, `/import_podcasts_opml_do`, `/podcasts`, `/radio`.
