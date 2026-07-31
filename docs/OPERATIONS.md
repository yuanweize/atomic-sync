# Operations, upgrade, and rollback

## Production checklist

Before deployment:

1. Use a versioned GHCR image pinned by digest.
2. Verify the Cosign signature and provenance.
3. Create a dedicated state directory owned by the container UID/GID.
4. Copy or mount a least-privilege `rclone.conf`; never include it in an image build context.
5. Bind the UI to loopback, Tailscale, or an authenticated reverse proxy.
6. Set a random `ATOMIC_API_TOKEN` of at least 32 bytes.
7. Mount the physical source branch read-only for the initial rollout.
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
6. **Move canary:** only after backup and final checks, temporarily permit source writes for one unit.
7. **Gradual expansion:** increase scope and concurrency slowly while monitoring Drive API quota, CIFS latency, and media playback.

Never enable write access and broad move jobs in the same change that first introduces the service.

## State backup

Stop API writes or stop the container, then copy all three SQLite files if present:

```bash
cp atomic-sync.db atomic-sync.db.backup
cp atomic-sync.db-wal atomic-sync.db-wal.backup 2>/dev/null || true
cp atomic-sync.db-shm atomic-sync.db-shm.backup 2>/dev/null || true
```

An online SQLite backup tool is preferred when zero downtime is required. Never restore only the main file while an unmatched WAL remains.

The remote `.atomic-sync-staging` directories are recovery material, not cache. Do not bulk-delete them during an incident.

## Upgrade

1. Read the release notes and database compatibility notes.
2. Back up the SQLite state.
3. Pull the new digest without recreating containers.
4. Validate the updated Compose model.
5. Recreate only `atomic-sync`.
6. Wait for `/api/ready` and inspect the first startup logs.
7. Confirm the UI version/commit and recent run states.

Do not run `docker compose up -d` without a service name in a busy shared stack. It may recreate unrelated services.

## Rollback

If health or logs regress:

1. Stop only the Atomic Sync service.
2. Restore the previous image digest.
3. Restore SQLite only if the release changed its schema incompatibly.
4. Start only Atomic Sync and verify `/api/ready`.
5. Leave all source and staging data untouched until run history is reconciled.

An interrupted non-terminal run becomes `failed` on restart. Review its source, staging, and final paths before retrying.

## Log review

Healthy startup contains one structured `atomic-sync listening` entry with version and commit. Investigate:

- `job run failed` or `unit failed`
- `final verification failed`
- `immutable merge failed`
- `published but source cleanup failed`
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
