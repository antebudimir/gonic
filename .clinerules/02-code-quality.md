# Code Quality

## Backend (Go)

- Idiomatic Go: return errors, wrap with context (`fmt.Errorf("...: %w", err)`), no panics in request paths.
- Keep handlers thin: parse params → authorize → query → map to `spec.*`. Shared logic belongs in a helper beside the handler.
- gorm v1 for all queries (`github.com/jinzhu/gorm`); raw SQL only inside migrations.
- Pass `r.Context()` down; never `context.Background()` inside a request path.
- Build Subsonic responses from `server/ctrlsubsonic/spec/`; add a field there, not an ad-hoc map.
- SQLite: keep the dialect patterns the neighbouring code uses — `COLLATE NOCASE` for names, `IN (?)` for batched lookups, chunked preloads (`db/db.go`).
- Table-driven tests with `testify/require`; name tests for behaviour, not implementation.

## Admin UI (Go templates)

- Templates live in `server/ctrladmin/adminui/pages/` plus `components.tmpl`; they are embedded with `//go:embed`, so rebuild to see a change.
- Handler data goes through `templateData` (`server/ctrladmin/ctrl.go`). Add a field there rather than passing maps.
- Match the markup and utility classes of the nearest existing page; the stylesheet is prebuilt (`04-ux.md`, `07-frontend.md`).

## Tests

- Add tests with the change: `*_test.go` next to the code; fixtures under `server/ctrlsubsonic/testdata/`.
- Test behaviour, not implementation: assert on responses and rendered output, not private fields.
- Derive expected values; never hardcode today's date, locale or timezone.
- Never point a test at a real database file: use `db.NewMock(":memory:", ...)` or a `t.TempDir()` (`06-backend.md`).

## Hygiene

- No dead code, no commented-out blocks, no TODO without a tracked reason.
- Keep fork code separable from upstream code: prefer new files and helpers over editing shared upstream functions in place, so cherry-picks stay clean.
