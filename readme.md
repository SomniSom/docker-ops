# Technical specification: **dq** — Docker Quick-ops

**Language:** English · [Русский](readme.ru.md)

**Status:** Go implementation (draft specification in this file).  
**CLI binary:** the **`dq`** executable in this repository; the Bash/Python wrapper has been removed.

**Full product name:** **Docker Quick-ops** (binary: `dq`).

---

## 1. Rationale and goals

### 1.1 Problem (historically)

- The previous stack based on **Bash**, **rsync/ssh/scp**, and **Python** for YAML led to drift between environments.

### 1.2 Goal

**A single statically linkable `dq` binary** that:

- reads configuration from `**docker-ops.yaml**` or `**docker-ops.yml**` only at the **project root** (no alternate paths such as `scripts/`);
- keeps secrets and sensitive values in a separate `**dq.env**` file (listed in `**.gitignore**`), merged into settings;
- exposes a CLI built with **spf13/cobra**;
- generates **shell completions** (bash/zsh/fish as needed) and **man pages** from Cobra metadata (or related tooling);
- transfers files over **SSH from Go** without requiring `rsync`/`scp` on the user machine;
- supports **cross-compilation** for major OS/arch combinations (linux/windows/darwin, amd64/arm64, etc.).

### 1.3 Non-functional expectations

- **Runtime self-sufficiency:** no Python and no mandatory `rsync` on the client.
- Predictable behaviour and clear error messages.
- Versioned releases (tags, changelog — as decided).

---

## 2. Terminology

| Term | Meaning |
| --- | --- |
| **Project root** | Directory from which the user runs `dq` (current `cwd`). |
| **Local mode** | No remote host configured, or remote explicitly disabled (`--local` / config flag / environment variable). |
| **Remote mode** | Host and server path are set; commands run **over SSH** on the remote machine (`dq` is **not** installed on the server). |
| **`source` deploy mode** | Sync project tree to the server + remote `reup` (build on server). |
| **`artifacts` deploy mode** | Server receives image-only compose, optional app config if set, image via registry or `save/load`. |

---

## 3. Naming and branding

- **Full name:** Docker Quick-ops.
- **Executable:** `dq`.
- Use **Docker Quick-ops** (`dq`) in `dq version`, man `NAME`, and documentation.

---

## 4. Solution architecture

### 4.1 Language and modules

- **Go** (version pinned in go.mod, LTS branch).
- **Cobra** — root command `dq`, subcommands instead of “first argv = command”.
- Packages (roughly): `config`, `compose` (CLI and API if needed), `remote/ssh`, tree mirroring, `deploy`, `internal/version`.

### 4.2 Configuration

- **Main file:** only `./docker-ops.yaml` or `./docker-ops.yml` at the **project root** (relative to cwd). No other standard paths (including `scripts/`).
- **Secrets:** separate `**dq.env**` at project root (dotenv / shell-assignments), **must** be in `.gitignore`; values merge into config (precedence vs YAML fields — defined in code; env file usually wins).
- Validation on startup; clear errors when required fields for a command are missing.

### 4.3 SSH and file copy

- Key-based auth (path from config / ssh-agent / OpenSSH defaults).
- **Directory sync** without `rsync` on the client: **SFTP + size/mtime comparison**; no repo size cap for now (may add later).
- **Image stream** `docker save | … | docker load`: over SSH via stdin or temp file on server — implementation-specific (disk and safety).

### 4.4 Docker interaction

- **Current:** only **`docker compose`** (**Compose V2** plugin for Docker CLI). Standalone **`docker-compose` (V1) is not supported** — if the plugin is missing, `dq` prints install hints (see Docker docs). Same locally and **on the remote host over SSH** (remote shell runs commands; **`dq` is not installed on the server**).
- **Extensions:** may use `**docker.sock**` (local Docker API).
- **Planned:** **HTTP Docker API** (including **remote** hosts). Access on the server via **SSH to the Docker socket** (`/var/run/docker.sock` on the remote host): forward/tunnel in the SSH session (local Unix socket or TCP proxy on the client), without exposing the Docker API on the network. Transport details — in implementation (see **§14.2**).

### 4.5 Remote execution (no `dq` on server)

- Subcommands such as `up`, `logs`, `deploy` in remote mode use **one SSH session / remote shell** running `**docker compose …**` in `remote_path`.
- In **`artifacts`** mode, **deploy files** are delivered (image compose, optional app config, etc.), **not** the `dq` binary.

### 4.6 Cobra: shell completion and man

- `dq completion bash|zsh|fish|powershell`.
- **`dq man [subcommand…]`** — man page from Cobra metadata, viewed with **`man -l`** (**man-db** / **groff**). If `man` is not in `PATH`, troff source goes to stdout.
- Static generation into **`man/man1/`**: **`make gen-man`** (`go run ./tools/genman`); install: **`make install-man`** (`MANPREFIX`, default `/usr/local/share/man`).
- Single source of truth: Cobra → `--help`, completion, and man.

---

## 5. Functional requirements

### 5.1 General behaviour

- Project root: **current working directory**; optional `--project-dir` / `**DQ_PROJECT_ROOT**`.
- Load `docker-ops.yaml` | `docker-ops.yml` + `dq.env` when present.
- Setting precedence: **§14.1**.

### 5.2 Configuration fields

Minimal set (snake_case in YAML):

| Area | Fields |
| --- | --- |
| Compose | `compose_project_name`, `compose_file`, `compose_service` |
| Remote | `remote_ssh`, `remote_path`, `ssh_identity` |
| Sync | `exclude` list (global); options equivalent to legacy `rsync_extra` map to internal sync |
| Deploy | `deploy_mode` (`source` or `artifacts`), `deploy_image`, `deploy_push`, `deploy_use_registry`, `deploy_save_load`, `deploy_save_compress` |
| Extra paths | `deploy_include` — relative to project root |
| Application | `**app_config**` (or equivalent): **optional** path to app config for copying in `artifacts` and for `config-check`; if unset — **check/copy not required** (often no file exists) |
| UX | `help_show_effective` and others as needed |

### 5.3 CLI commands

| Command | Locally | Remotely |
| --- | --- | --- |
| `help` / `--help` | yes | yes |
| `man` | man from Cobra (`man -l`) | yes |
| `env` (config template) | local only | do not proxy |
| `config-check` | if `app_config` set — verify file; else no-op or info | same over SSH |
| `build`, `pull`, `up`, `down`, `reup`, `ps`, `restart` | `docker compose …` | SSH + `docker compose …` in `remote_path` |
| `status` | ps + log tail | same over SSH |
| `logs`, `logs-tail`, `exec` | follow / tail / `exec -it` in terminal | SSH; PTY for follow and interactive `exec` |
| `deploy` | SFTP sync + local `docker` for `artifacts`; needs configured remote | `source` or `artifacts` |

**`deploy`:**

- **`source`:** directory on server, tree sync with exclude, copy **app_config** when configured and file exists, then remote `reup`.
- **`artifacts`:** optional local `docker build` (`deploy_push: true` in config or **`dq deploy --build`** without requiring `deploy_push`); registry vs save/load; deliver `docker-compose.image.yml`, **app_config** if set, `deploy_include`; on server over SSH: `config-check` (if applicable) → `up` or `pull`+`up`. **The `dq` binary is not copied to the server.** Draft `docker-compose.image.yml` from base compose: **`dq gen-image-compose`** (one service with `build:`, others `image:` only).
- **Data on server (`artifacts`):** the whole `remote_path` tree is **not** mirrored — arbitrary files and dirs (e.g. **`db-data`** for PostgreSQL) are **not** deleted unless you list them in **`deploy_include`** and overwrite from local. Ensure **`docker-compose.image.yml`** uses the same volume/path for DB data (`./db-data:/var/lib/postgresql/data`, etc.); plain **`docker compose up`** does not wipe a host bind-mount dir. **`source`** behaves differently (sync may delete extras on server) — riskier for a live DB inside the project tree.

### 5.4 `env` command (template)

- Print `docker-ops.yaml` template to stdout or `--output` with `--force`.
- `--anonymize`: do not substitute real `remote_ssh` / `remote_path` from config/env.

---

## 6. Platforms

- **Linux** — primary target.
- **Windows initially:** **WSL2** is enough (full native Windows not required in v1).

---

## 7. Build and distribution

- Go module: **`github.com/SomniSom/docker-ops`** (repo: https://github.com/SomniSom/docker-ops). Install from source: `go install github.com/SomniSom/docker-ops/cmd/dq@latest`.
- `go build -o dq` / `make build` — see **Makefile** (`VERSION`, `GIT_COMMIT`, ldflags → `internal/version`).
- **OS/arch matrix:** **GitHub Actions** (`.github/workflows/ci.yml`) — tests on Linux / macOS / Windows and cross-build `linux|darwin|windows` × `amd64|arm64`.
- **GoReleaser** (`.goreleaser.yaml`): same targets, archives (`tar.gz` / `zip` on Windows), **`checksums.txt`**. Locally: `make goreleaser-check`, snapshot without release: `make goreleaser-snapshot` ([goreleaser](https://goreleaser.com/install/) required). GitHub release: tag **`v*`** → **`.github/workflows/release.yml`**.
- `dq version` (git tag + commit).
- man — `make gen-man` / `make install-man` (**§4.6**).

---

## 8. Localization

- UI language (messages, subcommand help, errors): **English by default**.
- If the system locale is **Russian** (`LANGUAGE`, `LC_ALL`, `LC_MESSAGES`, `LANG` — `ru` prefix) — use **Russian** (`internal/locale`).
- Explicit: **`DQ_LANG=en|ru|auto`** (overrides auto) or global **`dq --lang en|ru|auto`** (persistent flag; e.g. `dq up --lang ru` allowed).
- Product name in `dq version` and root help: **Docker Quick-ops** (not translated).

---

## 9. SSH security

- SSH host key behaviour like **accept-new** for known_hosts (analogous to `StrictHostKeyChecking=accept-new` in OpenSSH).

---

## 10. Licensing

- Project license: **Apache License 2.0** (`**LICENSE**` in repo root). Fork and use per license terms.
- Binary signing (cosign, etc.) not required in v1.

---

## 11. Legacy file names

- Old names (`**docker-ops.remote.yaml**`, `**docker-ops.remote.env**`, etc.) are **not** supported: only `**docker-ops.yaml` / `docker-ops.yml**` + `**dq.env**`. Migrate manually per docs.

---

## 12. Migration from the Bash version

- Mapping old `docker-ops.remote.*` / env vars → `docker-ops.yaml` + `dq.env` — document in roadmap table.
- **No** automatic migration (optional tool separately).
- Bash/Python scripts **removed** from the repo; use **`dq`** only.

---

## 13. Acceptance criteria (draft)

- All commands from §5.3 in local mode on Linux with Docker installed.
- Remote mode without `dq` on server: e2e `deploy` (`source` and `artifacts` with save/load) over SSH.
- No Python or mandatory `rsync` on client; sync via SFTP + size/mtime.
- Config only at root: `docker-ops.yaml` | `docker-ops.yml`; secrets in `dq.env`; merge order **§14.1**.
- `dq completion bash|zsh` and man per §4.6.
- Localization: en default, ru for Russian locale / `DQ_LANG` / `--lang` (**§8**).
- WSL2 user scenario documented and verified where possible.

---

## 14. Decisions (clarifications)

### 14.1 Configuration precedence and `dq.env`

For each parameter (YAML key, logical env var name), order is **weaker → stronger** (stronger overrides):

1. `**docker-ops.yaml` / `docker-ops.yml**` — base values from file at project root.
2. `**dq.env**` — secrets file; **overrides** matching parameters from YAML (same key in either file; if both exist, **`dq.env` wins**).
3. **Process environment** (CI/CD, `export`, systemd) — **highest**; overrides YAML and `dq.env`.

Parameters only in YAML and absent from `dq.env` and process env come from YAML. Empty or missing lines in `dq.env` should not wipe YAML without an explicit rule in code (recommendation: treat empty as “unset” and do not override YAML).

### 14.2 Remote Docker API and SSH socket

Future Docker API access on remote hosts: **via SSH to the Docker socket** on the server (typically Unix socket `/var/run/docker.sock`). `dq` should establish an **SSH tunnel** (or library equivalent) so the user machine gets a local endpoint (Unix socket or `127.0.0.1:port`) pointing at remote Docker; API client talks to that endpoint. No need to expose the Docker daemon on the internet.

### 14.3 License

**Apache License, Version 2.0**; full text in **`LICENSE`** at repo root.

---

## Implementation (Go)

- Sources: `cmd/dq`, `internal/config`, `internal/cli`, `internal/compose`, `internal/deploy`, `internal/sshexec`, `internal/remote`, `internal/version`, `tools/genman`.
- **No** Bash/Python scripts in the repo (formerly `scripts/`).
- Build: `make build` → `bin/dq`; `make install` via `go install`.
- Tests: `make test` / `make test-unit` (`-race`); `make test-integration` — `-tags=integration` (Docker with **Compose V2** plugin required).
- Detailed status — **Roadmap** below.

---

## Roadmap (implementation status)

Legend: `[x]` done · `[ ]` not done / planned.

### Configuration and validation

- [x] `docker-ops.yaml` / `docker-ops.yml` only at project root
- [x] `dq.env` and merge order **§14.1** (YAML → dq.env → process env)
- [x] Defaults `compose_project_name` (directory name), `compose_file`, `compose_service`
- [x] `dq validate` (YAML, `deploy_mode`, `app_config` on disk, `dq.env` syntax)
- [x] Clear errors on YAML syntax (line context, indentation hints)
- [x] `dq env` (`--output`, `--force`, `--anonymize`)
- [ ] Deeper semantic validation (e.g. all deploy fields, `ssh_identity` exists)

### Local mode (Docker Compose via CLI)

- [x] `build`, `pull`, `up`, `down`, `reup`, `ps`, `restart`, `exec`, `status`, `logs`, `logs-tail`
- [x] Check `app_config` before `up` / `reup` when path set
- [x] Integration test with Docker (`-tags=integration`)
- [x] **Compose V2** required (`docker compose`); no `docker-compose` V1 fallback — missing plugin shows install hint

### Remote mode (no `dq` binary on server)

- [x] Compose commands over **SSH** in `remote_path` (`internal/sshexec`, `internal/remote`) — `build`, `pull`, `up`, `down`, `reup`, `ps`, `restart`, `exec`, `status`, `logs`, `logs-tail`
- [x] Auth: `ssh_identity` (path, `~` allowed) and/or **ssh-agent** (`SSH_AUTH_SOCK`)
- [x] New host keys: append to `~/.ssh/known_hosts` (**accept-new**); key change — refuse
- [x] TTY / PTY for `logs -f` with interactive terminal
- [x] Disable remote: `DOCKER_OPS_USE_REMOTE=0` or `use_remote: false` in YAML

### Deploy

- [x] **`deploy` in `source` mode:** SFTP tree sync, exclude list, `deploy_include`, optional `app_config`, remote `reup`
- [x] **`deploy` in `artifacts` mode:** `docker build`, registry vs `save`/`load`, deliver `docker-compose.image.yml` and files, remote `pull`+`up` or `up`
- [ ] Map `rsync_extra` → internal sync options (if still in spec)

### Docker beyond `docker compose` CLI

- [ ] Local **docker.sock** / API (**§4.4**)
- [ ] Remote Docker API via **SSH tunnel to socket** (**§14.2**)

### CLI, docs, release

- [x] Cobra, `dq completion` (bash / zsh / fish / powershell)
- [x] `dq version` + Makefile ldflags
- [x] **Man pages:** `dq man`, `make gen-man` / `make install-man` (**§4.6**)
- [x] **GoReleaser** + CI OS/arch matrix, archives, checksums (**§7**, `.goreleaser.yaml`, `.github/workflows/`)
- [ ] Short **WSL2** doc (**§6**)
- [ ] Migration table old `docker-ops.remote.*` → `docker-ops.yaml` / `dq.env` (**§12**)

### Localization

- [x] Messages and help: **en default**, **ru** for Russian locale (**§8**), package `internal/locale`
- [x] **`--lang`**, **`DQ_LANG`**

### Other

- [ ] Artifact signing (cosign, etc.) — not required in v1 (**§10**)

---

*May move to `docs/spec-dq.md`; this `readme.md` is the living English spec. Russian: [readme.ru.md](readme.ru.md).*
