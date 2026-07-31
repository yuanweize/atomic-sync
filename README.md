<div align="center">
  <img src="docs/assets/atomic-sync-wordmark.svg" width="560" alt="Atomic Sync">

  <p><strong>Branch-aware, fail-closed media archive orchestration powered by rclone.</strong></p>

  <p>
    <a href="README.zh-CN.md">简体中文</a> ·
    <a href="docs/ARCHITECTURE.md">Architecture</a> ·
    <a href="docs/OPERATIONS.md">Operations</a> ·
    <a href="SECURITY.md">Security</a>
  </p>

  [![CI](https://github.com/yuanweize/atomic-sync/actions/workflows/ci.yml/badge.svg)](https://github.com/yuanweize/atomic-sync/actions/workflows/ci.yml)
  [![CodeQL](https://github.com/yuanweize/atomic-sync/actions/workflows/codeql.yml/badge.svg)](https://github.com/yuanweize/atomic-sync/actions/workflows/codeql.yml)
  [![Release](https://img.shields.io/github/v/release/yuanweize/atomic-sync?display_name=tag)](https://github.com/yuanweize/atomic-sync/releases)
  [![Container](https://img.shields.io/badge/GHCR-multi--arch-2496ED?logo=docker&logoColor=white)](https://github.com/yuanweize/atomic-sync/pkgs/container/atomic-sync)
  [![License](https://img.shields.io/github/license/yuanweize/atomic-sync)](LICENSE)
</div>

---

Atomic Sync copies a movie directory, a complete show, or one season as a single migration unit. It stages the unit, verifies it, publishes it, and verifies the final destination again. **Version 0.1.x is copy-only:** it never deletes source data, and both `mode: move` and `deleteSource: true` are rejected by the API and Runner.

It also understands the part a union filesystem hides: the same folder name on two mergerfs branches does **not** prove that the archive is complete. Atomic Sync compares the physical source and destination inventories and reports what is actually archived, partial, pending, conflicting, or empty.

## Why it exists

A command such as `rclone move --min-age 30d` evaluates age per file. A late subtitle, poster, or episode can remain behind while older files move first, splitting one media unit across storage providers.

Atomic Sync evaluates the newest file in the entire unit. A 30-day stable window means the complete directory must remain unchanged for 30 days before it becomes eligible. Eligibility does not authorize deletion: source cleanup is an external, operator-controlled procedure performed only after writers have been stopped.

```mermaid
flowchart LR
  A[Discover stable directory unit] --> B[Validate fixed hierarchy]
  B --> C[Pin destination]
  C --> D[Copy to hidden staging]
  D --> E[Exact bidirectional staging check]
  E --> F{Destination exists?}
  F -->|No| G[Promote staging]
  F -->|Yes, fail policy| X[Stop; retain source and staging]
  F -->|Immutable merge| H[Copy missing files; never overwrite]
  G --> I[Verify source is present at final]
  H --> I
  I --> J[Complete; source retained]
```

## What makes it safe

- **Dry-run by default.** Omitted API values and the UI both create a dry run.
- **Copy-only in v0.1.x.** The API and Runner reject move mode and `deleteSource`; the official image does not contain rclone's `purge` command.
- **Fail-closed conflicts.** Existing destination units stop publication unless `merge-immutable` is explicitly selected.
- **Immutable merge.** Missing files may be added, but an existing different file is never overwritten.
- **Directory-only units.** Execution requires a directory at one fixed grouping depth. Shallow files and parent/child unit overlap stop the run before publication.
- **Two verification gates.** Staging must match the complete source unit exactly in both directions. A new destination must also match exactly; only `merge-immutable` uses a one-way final check so reviewed destination-only content can remain.
- **No shell interpolation.** rclone is launched with an argument vector, not a shell command.
- **Deterministic placement.** A unit is pinned to one weighted destination in SQLite and stays there across retries.
- **Cross-job isolation.** Equal or nested source/destination paths cannot be configured in separate jobs.
- **Lifecycle-aware shutdown.** SIGTERM cancels active work, waits for workers, and marks interrupted records failed on restart while preserving staging.
- **Hardened container.** UID 1000, read-only root filesystem, zero Linux capabilities, no-new-privileges, and a dedicated state path.
- **Session-only UI token.** The SPA is public to load; protected APIs require a constant-time Bearer-token check. The token stays in `sessionStorage`.

## Branch-aware archive status

Atomic Sync lists each physical branch once and compares files inside every atomic unit by relative path and size. This is intentionally metadata-first so a dashboard scan does not read terabytes from a CIFS source. An execution still performs the configured checksum or size verification.

| Status | Meaning | Safe next action |
|---|---|---|
| `archived` | Destination has content; the source has no files (an empty directory shell may remain) | Confirm mount health and retain the audit record |
| `ready-to-verify` | Every source path and size exists at the destination, but the source still exists | Quiesce writers, run independent final verification, then follow the manual cleanup procedure |
| `partial` | The destination has the unit but is missing some source files | Immutable merge or investigate |
| `pending` | Source has files and the destination unit is absent | Archive candidate |
| `conflict` | A relative path has a different size/type, or files exist outside the assigned destination branch | Stop; select the authoritative copy and branch |
| `empty` | Only an empty directory shell is visible | Review or ignore |

`partial` also covers a split unit whose two branches contain entirely complementary files: it can legitimately show 0% source coverage while GD already contains other files for that movie or show. An empty destination directory alone still counts as `pending`, not partial or archived.

See [Archive analysis](docs/ARCHIVE-ANALYSIS.md) for mergerfs examples and the exact decision rules.

`archived` is inferred from the current physical inventories; it is not historical proof of a completed run. A successful empty source listing is valid after a complete archive, so verify the physical mount before analysis. The reference Compose file refuses to create missing bind sources, but it cannot distinguish a healthy empty share from an existing mountpoint whose filesystem is offline.

### Verification modes

- `verify: checksum` runs `rclone check --download`. It reads the full contents of every compared file from both endpoints, so it is backend-independent but can generate substantial CIFS, network, and Drive I/O.
- `verify: size` runs the size-only comparison. It checks paths and byte counts without reading file contents and therefore provides weaker assurance.

The source-to-staging gate is bidirectional and exact for the complete directory unit: an extra or missing staging object blocks publication. A newly created destination receives the same exact bidirectional check. Only `merge-immutable` uses a source-to-destination one-way final check, allowing reviewed destination-only posters, subtitles, or previously archived files to remain.

## Quick start

Requirements: Docker Engine with Compose v2 and an existing `rclone.conf`.

```bash
git clone https://github.com/yuanweize/atomic-sync.git
cd atomic-sync
mkdir -p data rclone source
cp /path/to/rclone.conf rclone/rclone.conf
cp .env.example .env
```

Generate a token and place it in `.env`:

```bash
openssl rand -hex 32
```

Ensure the dedicated state and rclone config directory are owned by the image's fixed UID/GID 1000. Rclone refreshes OAuth tokens with a temporary file plus atomic rename, so this dedicated directory must be writable. Never recursively change ownership of a shared media root.

```bash
sudo chown -R 1000:1000 data
sudo chown -R 1000:1000 rclone
sudo chmod 700 rclone
sudo chmod 600 rclone/rclone.conf
docker compose -f compose.yaml -f compose.dev.yaml up -d --build
docker compose ps
```

Open `http://127.0.0.1:8088`, enter the API token, and create a **dry-run copy** job first. The default source mount is `/sources/media` and is read-only. The NUE v0.1.x rollout keeps `/data/storagebox/media` mounted read-only inside Atomic Sync and runs separate movie/TV dry-run jobs before any copy canary.

### Production image

Release tags publish signed `linux/amd64` and `linux/arm64` images with SBOM and provenance attestations:

```yaml
services:
  atomic-sync:
    image: ghcr.io/yuanweize/atomic-sync:0.1.3@sha256:<release-digest>
```

Pin the digest from the release before deployment. Do not use `build: .` when embedding the service into an unrelated Compose project.

The official image deliberately links only rclone's `local`, `drive`, and
`crypt` backends and only the commands Atomic Sync executes. This covers the
StorageBox/CIFS → Google Drive deployment while reducing the runtime attack
surface. Build a reviewed custom image if another rclone backend is required.

Each GitHub Release includes `image-digest.txt`, Linux binaries, and `SHA256SUMS`. The workflow requires the GHCR package to be public, fails closed if anonymous access is unavailable, checks per-platform SBOM/provenance and vulnerabilities, then signs and immediately verifies the immutable digest with GitHub OIDC. A first publication may require setting package visibility to Public in GitHub Packages and rerunning the failed workflow. Verify a release before use:

```bash
IMAGE="$(cat image-digest.txt)"
docker pull "$IMAGE"
cosign verify "$IMAGE" \
  --certificate-identity-regexp '^https://github.com/yuanweize/atomic-sync/.github/workflows/release.yml@refs/tags/v[0-9]+\.[0-9]+\.[0-9]+$' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
sha256sum -c SHA256SUMS
```

## Configuration

| Variable | Default | Purpose |
|---|---:|---|
| `ATOMIC_LISTEN` | `127.0.0.1:8080` | HTTP listen address; the container explicitly sets `:8080` |
| `ATOMIC_DATA_DIR` | `/data` | SQLite and durable state directory |
| `ATOMIC_API_TOKEN` | empty | Bearer token; if set it must be at least 32 characters, and non-loopback listeners require it |
| `ATOMIC_RCLONE_BIN` | `rclone` | rclone executable path |
| `ATOMIC_MAX_CONCURRENCY` | `2` | Global concurrent rclone-process ceiling |
| `ATOMIC_RCLONE_TRANSFERS` | `2` | Maximum parallel transfers inside each rclone process |
| `ATOMIC_RCLONE_CHECKERS` | `2` | Maximum parallel checks inside each rclone process |
| `ATOMIC_RCLONE_TPS_LIMIT` | `2` | Per-process backend transactions per second; burst is fixed at 1 |
| `ATOMIC_LOG_FORMAT` | `json` | `json` or text structured logs |
| `RCLONE_CONFIG` | `/config/rclone/rclone.conf` | Explicit rclone configuration path |
| `RCLONE_CONFIG_DIR` | `./rclone` | Host path used by Compose for the dedicated writable config directory |

The application never writes `.env` or edits rclone configuration, but its rclone child may atomically persist refreshed OAuth tokens in the dedicated `/config/rclone` bind. Keep that directory private (`0700`) and `rclone.conf` at `0600`; the container root and media source mounts remain read-only. Jobs, assignments, runs, and branch analyses are stored in `atomic-sync.db`. Local sources are restricted to `/sources/...`, local destinations to `/destinations/...`, and remote sources are rejected. Jobs always copy a complete directory unit; include/exclude filters are rejected in v0.1.x.

## Guarantees and boundaries

Atomic Sync provides a staged, verified copy-publication protocol. Object-storage directory operations are not ACID transactions, and a destination-side `moveto` may be implemented as multiple object operations. The source is always retained by v0.1.x.

`merge-immutable` is deliberately not fully atomic: it can add missing objects before a later conflict is discovered. It never overwrites a different destination object. Its hidden staging copy is retained even after success as recovery and audit material; Atomic Sync never issues an automatic staging cleanup. For a new destination, promotion moves the destination-side staging directory into its final name, so there is no separate staging copy to clean.

Source deletion is outside the v0.1.x trust boundary. Stop Sonarr/Radarr importers and every other writer for the selected unit, independently verify the final destination, take or confirm a recovery copy, delete only that reviewed source directory with an external administrative tool, and rescan before resuming writes. See [Operations](docs/OPERATIONS.md#manual-source-cleanup-outside-atomic-sync).

SQLite makes one Atomic Sync instance the supported topology. Do not run multiple replicas against the same database.

## Documentation

- [Architecture and state machine](docs/ARCHITECTURE.md)
- [Branch-aware archive analysis](docs/ARCHIVE-ANALYSIS.md)
- [Production operations and rollback](docs/OPERATIONS.md)
- [HTTP API](docs/API.md)
- [Threat model](docs/SECURITY-MODEL.md)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)

## Development

Go 1.25 or newer is required.

```bash
gofmt -w .
go vet ./...
go test -race -cover ./...
ATOMIC_API_TOKEN="$(openssl rand -hex 32)" docker compose config
docker build -t atomic-sync:dev .
```

The critical engine, API, model, store, and configuration packages are covered by unit and integration-style tests, including copy-only enforcement, exact staging verification, fail-closed hierarchy/conflict handling, immutable merge, shutdown, authentication, and mergerfs branch-status scenarios.

## Roadmap

- Scheduled polling and filesystem watch adapters
- Checksum-on-demand for selected analysis units
- Resume and guided cleanup for retained staging
- Prometheus metrics, notifications, and fine-grained RBAC
- Multi-node agents for remote NAS push workflows

Atomic Sync is released under the [MIT License](LICENSE).
