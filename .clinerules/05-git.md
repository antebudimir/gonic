# Git

- `master` is the only long-lived branch and must always be deployable.
- Never force-push `master`; never rewrite history that is already pushed.
- Take a backup branch before risky history work: `git branch backup/<what>-$(date +%F)` (e.g. `backup/pre-upstream-sync-2026-11-11`).
- Keep the backup branch until CI is green and the smoke test passes; delete it deliberately, not by accident.
- Run git write commands serially; never batch `git add`/`commit`/`rebase` in parallel.
- After any history rewrite or multi-command git operation, verify actual repo state (`git log`, `git show`) before acting on cached status output.
- Upstream work is cherry-picked, never merged — see `08-upstream-sync.md`. Never use GitHub's "Sync fork" button.
- Cherry-pick with `-x` so the `(cherry picked from commit …)` trailer lands in the message.
- Fork commit subjects use Conventional Commits (`feat:`, `fix(scope):`, `chore:`, `docs:`, `test:`, `ci:`). Upstream subjects stay verbatim.
- Use `git --no-pager` / `GIT_PAGER=cat` in scripts and agent commands.
- Quote rev-range arguments: zsh expands an unquoted `^` as glob negation, so `upstream/master ^HEAD` silently becomes `upstream/master`. Use `'^HEAD'` or `HEAD..upstream/master`.
- Never commit generated files, secrets, build output or a `.db` file.
- Tags (`vX.Y.Z-fork.N`) trigger the image publish: only tag a commit that passed CI (`06-ci.md`).
- Push the single tag (`git push origin vX.Y.Z-fork.N`), never `--tags` — that would push upstream's tags into the fork.
