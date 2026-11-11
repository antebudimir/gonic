# Upstream Sync

Authoritative source for sync strategy. Other files point here rather than restating it.

## What

Hard fork of [sentriz/gonic](https://github.com/sentriz/gonic). Divergence is expected and permanent: fork commits are not in upstream, and upstream originals are not ancestors of `HEAD` once cherry-picked. Do not chase a zero diff; `git status` will report both "ahead" and "behind" forever.

## Strategy: cherry-pick, never merge

- Merging `upstream/master` would drag in upstream's whole history and reintroduce conflicts with fork features; this fork keeps a readable, feature-shaped history instead.
- Never use GitHub's "Sync fork" button: it merges upstream wholesale, which is the thing this fork avoids.
- Never add `merge=ours` attributes: cherry-pick uses the per-file merge machinery, so an `ours` driver would make picks silently discard upstream changes.

## Candidate list

```bash
git fetch upstream
git cherry -v upstream/master HEAD     # upstream first, HEAD second
```

- `- ` marks a commit whose patch is already in `HEAD`; `+ ` marks one that is not. Classify only the `+` set.
- `git cherry` matches by **patch-id**, and patch-ids change when a pick needed conflict resolution — which these rules mandate. So an already-applied commit can still show as `+`. The tool shrinks the false candidates; it does not eliminate triage. Check the diff before re-picking.

## Steps

1. `git branch backup/pre-upstream-sync-$(date +%F)` — keep it until CI is green and the smoke test passes.
2. `git fetch upstream`.
3. Build the candidate list with `git cherry -v upstream/master HEAD`; classify the `+` set:
   - **Safe** — no conflict expected.
   - **Needs review** — touches a file the fork also changed; resolve by hand.
   - **High risk** — driver/lockfile/migration swap; skip unless critical and smoke-tested.
   - **Skip** — removed features, docs for removed features, version gates the fork does not satisfy, code depending on a skipped commit.
4. Cherry-pick in dependency order with `-x`; resolve conflicts inside the same commit and note the resolution in the message body.
5. Run the full gate set (`03-workflow.md`), including `fork-check .`.
6. Smoke test on a throwaway loopback port so a running instance is never touched:
   ```bash
   CGO_ENABLED=0 go build -o /tmp/gonic ./cmd/gonic
   mkdir -p /tmp/m /tmp/c /tmp/p /tmp/pl
   /tmp/gonic -listen-addr 127.0.0.1:4799 -music-path /tmp/m -cache-path /tmp/c \
     -podcast-path /tmp/p -playlists-path /tmp/pl -db-path /tmp/gonic.db
   curl 'http://127.0.0.1:4799/rest/ping?u=admin&p=admin&v=1.16.1&c=smoke'
   ```
7. Only then tag and push.

## Protected features

Every sync must leave these intact: public share links (`server/ctrlshare/`, `server/ctrlsubsonic/handlers_share.go`), smart playlists (`playlist/rules.go`, `server/ctrlsubsonic/smart_playlist.go`), OPML import (`podcast/opml.go`), scan schedule (`scanner/schedule.go`), transcode cache admin, dark admin theme. `fork.conf` asserts the code exists; `fork-check .` is the enforcement.

## Verifying the merge

- `fork-check .` — feature files present, routes wired, key fixes still in place, no conflict markers.
- Overlap review: `git diff upstream/master HEAD -- <file>` on files both sides touched; it should show fork additions, not deletions of fork code.
- Semantic check, not textual: upstream refactors often land next to a fork hook. Confirm the fork's behaviour still happens (read the function), not just that the text survived.
- Re-check the fork routes listed in `06-backend.md`.

## Tag last, verify first

- Order: sync → backend gates → fork invariants → smoke → CI green → tag → push.
- `version.txt` bumps in the **same commit** that receives the tag, never afterwards.
- Pre-tag check: `git show <tag-candidate>:version.txt` must equal the tag name without the `v` and `-fork.N` suffix.
- Tag convention: `vX.Y.Z-fork.N`, incrementing `N` monotonically regardless of upstream's version.
- Tag pushes publish an image that CI has not tested (`06-ci.md`); the local gates are the only validation.
- After pushing, confirm the run actually started; re-run or re-tag if it did not.

## Pitfalls

- Cherry-picked commits carry different hashes than upstream's originals; `git merge-base --is-ancestor <upstream-hash> HEAD` is false by design.
- A skipped commit's symbols may be referenced by a picked one — inline or gate the reference in the same commit.
- Rebrand-sensitive picks (strings, docs, templates) must be fixed to this fork inside the pick, never left for later.
- Upstream `CHANGELOG.md` sections are release-please output; fork entries go under `## Unreleased`.
