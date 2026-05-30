# AGENTS.md — guidance for AI coding agents

Before changing **deploy**, **remote SSH**, or **Docker** integration in this repo:

1. Read [readme.md](readme.md) (spec) §5.3, §14.1, §14.2.
2. Read [docs/ai-operator.md](docs/ai-operator.md) (decision trees, must-ask questions, backward compatibility).

## Do

- Keep **default `deploy_engine: compose`** — existing configs must behave unchanged.
- Add new config keys as **optional** with safe defaults; document in readme **and** `docs/ai-operator.md`.
- Provide **fallback** to SSH + `docker compose` when API/tunnel/layer sync fails (`deploy_engine: auto`).
- Update locale strings in **both** `internal/locale/catalog_en.go` and `catalog_ru.go`.
- Run `go test ./...` after deploy-related changes.

## Do not

- Do not set `deploy_engine: api` in examples or generated configs unless the user explicitly asked for API-only deploy.
- Do not remove or rename existing YAML/env keys.
- Do not change `source` deploy or compose subcommands (`up`, `logs`, …) to Docker API in the same change unless scoped in the task.
- Do not edit `.cursor/plans/` plan files unless asked.

## Key packages

| Package | Role |
|---------|------|
| `internal/deploy` | `dq deploy`, artifacts/source, transfer, API apply |
| `internal/dockerapi` | Local/remote Docker client, SSH socket tunnel, socket auto-detect |
| `internal/config` | YAML + dq.env + env overlay |
| `internal/sshexec` | SSH dial, bash, pipe |
| `internal/remote` | Remote `docker compose` CLI |

## Tests

- Unit: `internal/dockerapi`, `internal/config`, `internal/deploy/*`
- No Docker daemon required for default unit tests
