# Contributing to kanshi-agent

Kanshi is built from three repos: [core](https://github.com/kanshi-dev/core), [agent](https://github.com/kanshi-dev/agent), and [dashboard](https://github.com/kanshi-dev/dashboard). Work is planned on the [Kanshi v1.0.0 project board](https://github.com/orgs/kanshi-dev/projects/1); the roadmap and priorities live there. Please pick up (or file) an issue before opening a PR.

## Workflow

1. **Start from an issue.** Use the issue templates; every issue states Overview, why it matters, and checkable acceptance criteria.
2. **Branch from `main`**, named `<type>/<short-description>`, e.g. `fix/graceful-shutdown`, `feat/net-collector`.
3. **Commit with [Conventional Commits](https://www.conventionalcommits.org/):** `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`, `ci:`. Keep commits scoped and messages in the imperative ("fix reconnect loop", not "fixed").
4. **Open a PR** using the PR template and link the issue (`Closes #N`).
5. **CI must pass before merge** (build, vet, test). Do not merge red. Do not push directly to `main`.

## Rules of the repo

- **Bug fixes land with a regression test** that reproduces the bug: it fails before the fix and passes after.
- The agent runs on other people's servers: every blocking call must respect `context.Context`, and memory use must stay bounded under failure (see the batch buffer policy in `internal/pipeline/`).
- Collectors implement the `collect.Collector` interface and are registered in `internal/registry/registry.go`; metric points carry tags rather than encoding metadata in names.
- Standard Go `internal/` layout; no global state.

## Development

```bash
# core must be running (see kanshi-dev/core), then:
go run ./cmd/agent
go build ./... && go vet ./... && go test ./...   # what CI runs
```

Configuration is via `KANSHI_*` env vars (`KANSHI_CORE_ADDR`, `KANSHI_API_KEY`, ...). See `internal/config/`.

## Versioning

Semver from v1.0.0: bug fixes ship as `v1.0.x` patches, features wait for `v1.1.0`.
