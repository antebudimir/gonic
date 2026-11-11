# Workflow

## Gates, in order

```bash
# 1. compile, vet, tests (CGO-free build, same as the Dockerfile)
CGO_ENABLED=0 go build ./...
go vet ./...
go test ./...

# 2. lint (same version as CI)
golangci-lint run --config .golangci.yml

# 3. fork invariants
fork-check .
```

- Run the narrowest failing test first (`go test ./server/ctrlsubsonic -run TestName`), then the full suite before push.
- No Go toolchain locally: say so and run the gate in a container (`podman run --rm -v "$PWD:/src:z" -v gonic-gomodcache:/go/pkg/mod -w /src -e CGO_ENABLED=0 golang:1.27-alpine3.24 sh -c 'go build ./... && go test ./...'`) or wait for CI.
- CI shape (`06-ci.md`): `release.yaml` runs lint+test on master and on tags; `nightly-release.yaml` gates its image on the same test job; `test.yaml` only fires on `develop`.

## Commits

- One logical change per commit. A cherry-pick stays one commit; fold its fork fallout into it (`08-upstream-sync.md`).
- Prefix fork commits: `feat:`, `fix:`, `fix(area):`, `chore:`, `docs:`, `test:`, `ci:`. Upstream subjects stay verbatim.
- Update `README.md` in the same commit when behaviour, flags or env vars change.
- Update `CHANGELOG.md` under `## Unreleased` for user-visible changes; internal refactors and test-only changes do not.

## Never commit

- Generated build output other than the embedded stylesheet (which is committed by design — regenerate it with its source, `07-frontend.md`).
- Secrets or data: `*.db`, real music/podcast paths, tokens, session keys.
- Anything you have not run the gates against.

## When stuck

- Reproduce with the smallest command, then read the actual error.
- Check whether the behaviour is a fork feature or an upstream one (`AGENTS.md`).
- Report the blocker with the exact command and output instead of guessing.
