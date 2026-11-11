# Core Behavior

Applies to every task. Fork invariants: `06`–`08`. Gates: `03-workflow.md`.

## Orient first

- Read `AGENTS.md` before your first edit: what this fork added matters as much as what upstream owns.
- Read the target file and its callers; grep for a sibling handler before inventing a pattern.
- Check `git log --oneline -15` and `git status` before touching anything.
- State the plan when a task spans more than one file.

## Change minimally

- Make the smallest diff that solves the problem. No drive-by refactors, renames or reformatting.
- Match the surrounding code: naming, error handling, handler shape, test layout.
- Reuse existing helpers; do not add a second way to do the same thing.
- Do not add dependencies, flags, env vars or API surface without a concrete need.
- Never hand-edit generated output: `server/ctrladmin/adminui/static/style.css` is tailwind build output, `go.sum` is tool-managed.

## Verify

- Prove behaviour with a test or real command output, not by reading code.
- Run the gates in `03-workflow.md`; if a gate cannot run here, say so explicitly.
- Fix a failing test or explain precisely why it is unrelated — never leave the suite red.

## Guardrails

- Never commit secrets, tokens, `.db` files or real library data.
- Ask before irreversible calls: schema or migration changes, Subsonic response shape, image names, tags.
- Never publish to upstream's namespace (`sentriz/*`); this fork publishes to `ghcr.io/antebudimir/gonic` (`06-ci.md`).
- Never add `merge=ours` attributes to files here: cherry-pick uses per-file merge machinery, so `ours` would silently discard upstream changes (`08-upstream-sync.md`).
- Do not delete fork features to satisfy a request; propose a fork-shaped alternative instead.
