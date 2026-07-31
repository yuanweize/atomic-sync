# Operations, upgrade, and rollback

## Production checklist

Before deployment:

1. Use a versioned GHCR image pinned by digest.
2. Verify the Cosign signature and provenance.
3. Create a dedicated state directory owned by the container UID/GID.
4. Copy `rclone.conf` into a dedicated `0700` directory owned by the container UID/GID, with the file at `0600`; mount only that directory read-write so rclone can atomically persist OAuth token refreshes. Never include it in an image build context.
5. Bind the UI to loopback, Tailscale, or an authenticated reverse proxy.
6. Set a random `ATOMIC_API_TOKEN` of at least 32 bytes.
7. Mount the physical source branch read-only. Version 0.1.x is copy-only and never requires source write access.
8. Do not mount an entire host `/data` tree when only one source path is required.
9. Verify that each physical source is mounted with the expected filesystem type (`findmnt --target /path/to/source`) before starting the container or running analysis. The Compose bind uses `create_host_path: false`, but an offline filesystem can leave an existing empty mountpoint behind.
10. Confirm the official image's local/Drive/crypt backend set matches the deployment.
11. Keep JSON logs bounded with rotation.
12. Record the pre-deployment Compose hash and container states.

## First rollout

Use these phases:

1. **Health-only:** start the container with state and rclone config, but a read-only source mount.
2. **Branch analysis:** create separate movie and TV jobs; run analysis and review conflicts.
3. **Dry-run:** keep `dryRun: true` and confirm discovered unit boundaries.
4. **Copy canary:** copy one small, stable, source-only unit with `conflictPolicy: fail`.
5. **Immutable-merge canary:** use one reviewed partial unit and confirm destination extras remain untouched.
6. **Gradual copy expansion:** increase scope and concurrency slowly while monitoring Drive API quota, CIFS latency, and media playback.

For a quota-sensitive first canary, set all four concurrency controls to `1`: `ATOMIC_MAX_CONCURRENCY`, `ATOMIC_RCLONE_TRANSFERS`, `ATOMIC_RCLONE_CHECKERS`, and `ATOMIC_RCLONE_TPS_LIMIT`. Use size verification first, then schedule full-content verification only after Drive quota remains healthy.

Do not make the source writable to Atomic Sync. The API and Runner reject `mode: move` and `deleteSource: true`, and the official image omits rclone's `purge` command.

### NUE v0.1.x rollout

The current GD remote still uses rclone's shared Google Drive OAuth client. Rclone warns that this shared client is being retired during 2026. Atomic Sync reports that exact notice as an operational warning while continuing the inventory; all other or mixed `lsjson` diagnostics still fail closed. Configure a dedicated Google Drive OAuth `client_id` and `client_secret` before the shared client is retired. Mount the dedicated Atomic Sync rclone config directory read-write so token refresh can use a temporary file and atomic rename; do not reuse or expose the host mount service's config directory.

For NUE, bind `/data/storagebox/media` read-only inside the Atomic Sync container and keep the mergerfs union out of the execution path. Create separate movie and TV jobs, both as `copy + dryRun`, and run branch analysis serially while Drive quota is healthy. Recreate only the Atomic Sync service; do not restart or recreate Sonarr, Radarr, mergerfs, or the rclone mount. A later copy canary still uses the read-only source mount.

## Verification cost and semantics

- `verify: checksum` invokes `rclone check --download`. It reads every compared file in full from the source and destination, so plan for CIFS, network, Drive download, and local I/O pressure.
- `verify: size` compares paths and sizes without reading file contents. It is faster and cheaper but cannot detect equal-size corruption.

The staging gate is bidirectional and exact for the complete directory unit: staging may contain neither missing nor extra objects. A new destination's final gate is also exact. Only `merge-immutable` uses a one-way final gate that allows reviewed destination-only content.

Executable units must be directories at a single configured depth. Treat a shallow file or parent/child unit overlap as a library-layout incident; repair the structure and rerun discovery rather than widening the job.

## State backup

Stop API writes or stop the container, then copy all three SQLite files if present:

```bash
cp atomic-sync.db atomic-sync.db.backup
cp atomic-sync.db-wal atomic-sync.db-wal.backup 2>/dev/null || true
cp atomic-sync.db-shm atomic-sync.db-shm.backup 2>/dev/null || true
```

An online SQLite backup tool is preferred when zero downtime is required. Never restore only the main file while an unmatched WAL remains.

Remote `.atomic-sync-staging` directories are recovery material, not cache. `merge-immutable` deliberately retains its hidden staging copy even after success, and Atomic Sync never cleans it automatically. Promotion to a new destination consumes its staging directory by moving it to the final name. Inventory retained staging, associate it with completed/failed run IDs, and remove it only through a separately reviewed external procedure.

## Manual source cleanup outside Atomic Sync

Atomic Sync v0.1.x never deletes source files. If an operator later decides to reclaim StorageBox space, perform cleanup outside the application and one reviewed directory at a time:

1. Select a completed unit with a healthy physical source mount and a verified final destination. Do not rely on the mergerfs view or an `archived`/`ready-to-verify` label alone.
2. Stop every writer that can modify that unit, including Sonarr/Radarr imports, download post-processing, repair scripts, and manual uploads. Confirm there are no active file handles or in-flight transfers.
3. Record the source and destination inventories and confirm a recoverable backup or snapshot exists.
4. Run an independent exact content verification. `rclone check --download <source-unit> <destination-unit>` reads both copies in full; if operational limits require size-only verification, record that weaker decision explicitly.
5. Re-list the source after verification. If any path, size, or modification time changed, abort and restart the process after the unit is quiescent.
6. Delete only the reviewed physical source directory with an external administrative tool. Never delete through `/data/merged`, never use a library-root wildcard, and never bulk-clean retained staging in the same change.
7. Confirm the destination remains readable through the mergerfs consumer path, trigger the appropriate Sonarr/Radarr rescan, inspect logs, and only then resume writers.

This manual procedure is intentionally outside Atomic Sync's API, container permissions, and v0.1.x guarantee. It requires an operator to manage the remaining verification-to-deletion race while the source is quiescent.

## Upgrade

1. Read the release notes and database compatibility notes.
2. Back up the SQLite state and the dedicated rclone config directory without replacing a newer refreshed token with an older copy.
3. When upgrading from v0.1.2 or earlier, replace `RCLONE_CONFIG_PATH=<file>` with `RCLONE_CONFIG_DIR=<dedicated-directory>`, set the directory to UID/GID 1000 and `0700`, and set `rclone.conf` to `0600`. The old variable no longer controls the Compose mount.
4. Pull the new digest without recreating containers.
5. Validate the updated Compose model and confirm only `/data` and `/config/rclone` are writable binds while the root filesystem and source remain read-only.
6. Recreate only `atomic-sync`.
7. Wait for `/api/ready` and inspect the first startup logs.
8. Confirm the UI version/commit and recent run states.

Do not run `docker compose up -d` without a service name in a busy shared stack. It may recreate unrelated services.

## Rollback

If health or logs regress:

1. Stop only the Atomic Sync service.
2. Restore the previous image digest.
3. Preserve the current dedicated rclone config directory; do not overwrite a token refreshed after the backup.
4. Restore SQLite only if the release changed its schema incompatibly.
5. Start only Atomic Sync and verify `/api/ready`.
6. Leave all source and staging data untouched until run history is reconciled.

An interrupted non-terminal run becomes `failed` on restart. Review its source, staging, and final paths before retrying.

## Log review

Healthy startup contains one structured `atomic-sync listening` entry with version and commit. Investigate:

- `job run failed` or `unit failed`
- `invalid job configuration` for unsupported move/source-deletion settings
- shallow directory-unit or parent/child overlap errors
- source-to-staging exact verification failures
- `final verification failed`
- `immutable merge failed`
- repeated `context deadline exceeded`
- cloud `403` quota errors
- SQLite busy or disk-full errors

The application never logs the API token or rclone configuration. rclone error output is capped before being added to application logs.

## Drive quota pressure

When Drive returns query quota errors:

- Do not restart the rclone mount repeatedly.
- Pause new Atomic Sync executions and analyses.
- Identify media scanners or clients causing repeated directory traversal.
- Let the quota window recover.
- Resume with one analysis/job and low concurrency.

Atomic Sync branch analysis is serialized, but other applications using the same Drive project still share its quota.
