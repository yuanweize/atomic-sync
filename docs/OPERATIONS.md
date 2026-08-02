# Production operations, recovery, and rollback

Atomic Sync invokes rclone once against the assigned final destination. This maximizes throughput, but it also means operators must treat real move jobs as destructive cross-provider operations rather than transactional directory renames.

## Production checklist

Before deployment:

1. Use a versioned GHCR image pinned by immutable digest.
2. Verify its Cosign identity, provenance, SBOM, and release checksums.
3. Create a dedicated state directory owned by container UID/GID 1000.
4. Copy `rclone.conf` into a dedicated `0700` directory owned by UID/GID 1000, with the file at `0600`. Mount only this directory read-write so rclone can persist OAuth refreshes through temporary-file rename.
5. Configure a dedicated Google Drive OAuth `client_id` and `client_secret`; do not depend on rclone's retiring shared client for production traffic.
6. Choose private ingress: loopback, Tailscale, or an authenticated reverse proxy. A process listening directly on a Tailscale IP is still non-loopback and must keep application authentication enabled.
7. Set a random `ATOMIC_API_TOKEN` of at least 32 characters. It is an Atomic Sync administrator secret, not a system or Tailscale password. Only a process whose own `ATOMIC_LISTEN` is explicitly loopback may omit it; the reference container listens non-loopback internally and therefore requires it even when its host port is published only on loopback or Tailscale.
8. Bind only the physical source branch, never the mergerfs union and never the complete host `/data` tree.
9. Keep the source bind read-only for initial rollout, all dry-runs, and copy-only deployments. A real move requires an explicit writable-source opt-in.
10. Verify the source mount's expected filesystem type and remote availability before every analysis or destructive run. `create_host_path: false` prevents Docker from creating a missing bind, but cannot detect an offline filesystem behind an existing empty mountpoint.
11. Confirm that the official image's local/Drive/crypt backend set covers the deployment. A host-mounted CIFS/SMB source uses rclone's local backend.
12. Keep JSON logs bounded and record the current Compose model, image digest, database backup, and container restart counts.

## Choosing the directory-unit boundary

Atomic Sync transfers regular files grouped below directory boundaries; media is an important hierarchy preset, not an engine limitation:

- `folder` treats each top-level directory below `job.source` as a general-purpose unit;
- `depth` selects an exact positive hierarchy depth for project trees, datasets, exports, build artifacts, or other nested layouts;
- `show` is the media-named equivalent of the one-level `folder` boundary;
- `season` selects a two-level `Show/Season` media boundary.

All payload files must be below the selected boundary. A loose file directly under `job.source` is shallow and stops the complete run before rclone writes; Atomic Sync currently does not transfer root-level loose files as independent units. For a `folder` job, expose a dedicated parent whose next-level children are the units. For `depth: 2`, use a shape such as `Org/Project/file.ext`.

The `show` and `season` presets only select one- and two-level boundaries; they do not parse names, rename media, or judge library completeness. Symbolic links and special files are unsupported. Empty directories are not transfer-manifest entries and are not guaranteed to survive, while ownership, permissions, extended attributes, and other POSIX metadata are outside the contract. Use a filesystem-aware backup tool when those properties matter.

## Safe rollout sequence

Use a narrow progression and stop at the first unexplained result:

1. **Health only:** start Atomic Sync with a read-only source and verify `/api/ready`, version, commit, and logs.
2. **Physical-branch analysis:** create one job per physical source tree, run analyses serially, and resolve every `conflict`. In the Media/NUE reference deployment, keep movie and TV jobs separate.
3. **Dry-run discovery:** keep the source read-only, `dryRun: true`, and review unit boundaries. Either move conflict/verification combination can be planned without source write permission.
4. **Copy canary:** run one small stable source-only unit with `mode: copy` and `conflictPolicy: fail`.
5. **Immutable-merge copy canary:** select one reviewed partial unit and prove that destination-only content remains untouched.
6. **Gradual copy expansion:** raise scope and throughput while monitoring provider quota, source latency, memory, and downstream consumer health.
7. **Optional move canary:** only after the preceding phases, stop writers, opt in to a narrow writable source bind, and run one explicitly confirmed unit.

The repository default stable window is 30 days (`2,592,000` seconds). A three-day value (`259,200`) is acceptable for a scoped dry-run/canary job whose discovered units are manually reviewed; do not change the repository or long-running production default to three days.

### Single-unit canary scope

Atomic Sync runs every eligible unit below `job.source`; it has no include filter or single-unit selector. To guarantee a one-unit `folder` canary, replace the broad media bind in a reviewed canary Compose model and expose the selected media directory as the **only child** of a dedicated container-side parent:

The scoping rule is general: mount exactly one reviewed unit as the only child below a dedicated source parent. The YAML below is the Media/NUE reference example; substitute your own physical source tree and downstream consumers for non-media workloads.

```yaml
services:
  atomic-sync:
    volumes:
      - type: bind
        source: /data/storagebox/media/movies/Selected Movie (2026)
        target: /sources/canary/Selected Movie (2026)
        read_only: true
        bind:
          create_host_path: false
```

Set `job.source: /sources/canary` and `grouping: folder`. Do **not** set `job.source` to `/sources/canary/Selected Movie (2026)`: its media files would sit above the required folder boundary and fail as shallow files. Keep this bind read-only for analysis, dry-run, and copy. Only after the real-move checklist may this exact unit bind become writable.

Compose volume merging is implementation-sensitive. Inspect the rendered model and require exactly one canary media bind; remove the broad `/sources/media` bind rather than leaving a second alias to the same physical tree. Pause or remove every job that references the broad alias before starting the canary, because container-path overlap validation cannot detect two bind paths that resolve to the same host files.

### Quota-sensitive canary profile

```text
ATOMIC_MAX_CONCURRENCY=1
ATOMIC_RCLONE_TRANSFERS=1
ATOMIC_RCLONE_CHECKERS=1
ATOMIC_RCLONE_TPS_LIMIT=1
job.concurrency=1
verify=size
dryRun=true
paused=true
```

This size-only profile validates control flow with minimal metadata work; it does not prove content identity.

### NUE dry-run canary

NUE keeps the production default at 30 days and uses three days only for the current scoped canary jobs. The canary contract is:

```text
mode=move
deleteSource=true
conflictPolicy=fail
verify=size
dryRun=true
paused=true
settleSeconds=259200
concurrency=1
```

Keep `/data/storagebox/media` read-only for this dry-run. Recreate only Atomic Sync when changing its image or configuration; do not restart or recreate Sonarr, Radarr, mergerfs, downloaders, or the existing rclone mount.

## Enabling real copy

Copy mode preserves the source and works with the reference read-only bind.

1. Confirm the physical source and destination mounts/remotes are healthy.
2. Run a fresh analysis and review the selected unit and assignment.
3. Keep `deleteSource: false` and change only `dryRun` to `false`.
4. Start with `conflictPolicy: fail` and one source-only unit.
5. Inspect the run result, destination inventory, Atomic Sync logs, and rclone errors.
6. Refresh the relevant downstream consumer only after the destination is visible through its normal path. For Media/NUE, this means the appropriate Sonarr/Radarr rescan.

Use `merge-immutable` only for reviewed `partial` or `ready-to-verify` units. It writes missing files directly and never overwrites a destination object. During move, every overlapping source path is deliberately retained; Atomic Sync reports the unit partial after moving the missing objects. `--ignore-existing` skips those paths, so neither `verify: checksum` nor `verify: size` proves their content. Those leftovers require independent content proof and explicit operator cleanup. Because this is a direct merge, successfully added files remain at the destination.

## Enabling real move

Move mode maps to direct rclone move semantics: rclone transfers to the final path and removes source objects it considers successfully transferred. There is no whole-directory rollback.

Complete all of these steps before granting source write access:

1. **Scope one unit.** Use the [single-unit canary bind](#single-unit-canary-scope) from the physical source branch, never a mergerfs union, and record its exact source/destination paths and assignment. For Media/NUE, this is the StorageBox branch rather than `/data/merged`.
2. **Quiesce writers.** Stop every importer, post-processor, repair/rename task, manual uploader, and other process that can touch the unit. For Media/NUE this includes Sonarr/Radarr imports and downloader post-processing. Check for open file handles and in-flight rclone/CIFS activity.
3. **Confirm stability and recovery.** Keep or record an independent backup/snapshot and capture source inventory. Atomic Sync revalidates the discovery fingerprint immediately before rclone, pins rclone to those paths with a temporary manifest, and verifies the final path-and-size inventory after every non-dry-run copy or move; move then checks source residue. These metadata gates narrow the race, but the stable window is still not a lock or backup. An in-place, equal-size rewrite that preserves its old modification time is outside their proof.
4. **Review branch state.** Resolve every size/type or wrong-destination conflict before the move.
5. **Set destructive intent consistently.** Use `mode: move`, `deleteSource: true`, `dryRun: false`, and a conservative concurrency. Choose `fail` for a destination that must be absent; use `merge-immutable` only for a reviewed partial unit. Prefer `checksum` for stronger production evidence, or record the weaker assurance when operational constraints require `size`.
6. **Make only the narrow source bind writable.** Do not expose the mergerfs union, other media roots, Docker socket, or host root.
7. **Validate the merged Compose model.** Confirm that the intended source target is the only source bind changed from read-only to read-write.
8. **Confirm the exact job name.** The UI asks for it; direct API clients send `X-Atomic-Confirm-Job: <exact job name>` when starting a non-dry-run move.
9. **Run one canary and wait.** Do not start a second move until the first run, branch analysis, downstream visibility, and logs are all understood.
10. **Restore least privilege.** When no real move is scheduled, return the source bind to read-only and recreate only Atomic Sync.

The required Compose change is intentionally not part of the safe default. In a local, reviewed production override, replace the source volume at target `/sources/media` with the same narrow host path and `read_only: false`, then inspect the rendered model:

```bash
docker compose -f compose.yaml -f ./reviewed-move-override.yaml config
```

The project does not ship or enable this override automatically. Do not assume sequence-merging behavior is correct from the file alone; inspect the rendered `services.atomic-sync.volumes` entry and verify that no unrelated source became writable.

## Direct-transfer recovery

An interrupted copy or move may leave a partial final directory. This is expected evidence, not a reason to delete either branch automatically.

1. Pause the job and keep every writer quiesced.
2. Preserve the SQLite database and current source/destination inventories.
3. Verify that both physical mounts/remotes are healthy; never infer a completed move from an empty mountpoint.
4. Run branch analysis. Remaining source files plus some destination content should appear as `partial`; incompatible same-path objects appear as `conflict`.
5. Read the failed run's rclone error before changing files. Quota, network, permission, immutable-conflict, and source-change failures require different responses.
6. Resolve conflicts by explicitly choosing the authoritative file. Never overwrite or delete through the mergerfs view.
7. Retry the same job only after the failure is understood. Its persisted assignment keeps the unit on the same destination.
8. After recovery, verify content through the downstream consumer path and trigger its appropriate refresh before resuming writers.

There is no automatic rollback of files already copied or moved successfully. Rclone's direct operations are chosen for speed and retry behavior; Atomic Sync supplies evidence and deterministic placement rather than fabricating a distributed filesystem transaction.

## Verification choices

- `verify: size` compares paths and byte counts without reading file content. It is fastest and appropriate for an initial canary, but cannot detect equal-size corruption.
- `verify: checksum` maps to rclone's `--checksum` transfer comparison and a hash common to both backends when available. For local/CIFS to Drive, this normally reads the source to calculate MD5 and compares Drive's stored hash rather than downloading Drive data. On move jobs, `--ignore-existing` skips destination-overlap paths, so checksum mode does not verify those retained source objects. Rclone owns transfer verification, retries, and resumability.
- Every copy or move uses a temporary `--files-from-raw` manifest containing exactly the discovery fingerprint's file paths. It is removed after the invocation and is control data only—not staging or another payload copy. With a positive stable window, rclone also receives `--min-age <seconds>s`.
- After every non-dry-run copy or move, Atomic Sync lists the final destination and requires each discovered file at the same path and size; move then checks source residue. This is a metadata-completeness gate against the pre-transfer fingerprint, not a second content verification or `rclone check`.
- `rclone check --download` is an operator-run deep audit for selected quiesced units. It reads full content from both sides and can materially affect CIFS, network, Drive quota, and playback.

Hash availability and remote consistency are backend properties. When no common hash exists, rclone can fall back to size comparison; do not claim checksum proof in that case.

## Throughput tuning

Atomic Sync does not choose between rclone, provider SDKs, `cp`, or `rsync`. Rclone remains the only transfer engine; tuning changes bounded concurrency and provider settings, not the data plane. The pinned file manifest lets rclone avoid another unconstrained source-tree discovery, and the final `lsjson` is metadata-only, so the safety closure adds no second payload transfer.

The effective transfer fan-out is approximately:

```text
concurrent rclone processes × transfers per rclone process
```

The defaults are intentionally moderate (`2` processes, `2` transfers, `2` checkers, TPS limit `2`). After a dedicated Google OAuth client is active:

1. Keep process/job concurrency at `1` and raise rclone transfers first.
2. Raise checkers only after listing and hash traffic remains quota-safe.
3. Increase the TPS limit gradually while watching `403 userRateLimitExceeded`, `429`, retries, and latency. Set it to `0` only after a dedicated OAuth client and measured provider headroom justify omitting the explicit cap.
4. Increase process concurrency last; it multiplies every per-process limit.
5. Monitor container memory when raising Drive chunk size. Chunk size and per-transfer buffers multiply with transfer count and must fit the container's memory limit.
6. Measure end-to-end bytes per second and error/retry rate. More concurrency is slower when it saturates CIFS, memory, or provider quota.

Rclone's shared Google Drive OAuth client is scheduled for retirement and shares quota with many users. Configure a private Google Cloud OAuth application before aggressive tuning; client credentials and refreshed tokens must remain only in the dedicated rclone config directory.

## State backup

Stop API writes or stop only the Atomic Sync container, then copy all SQLite files that exist:

```bash
cp atomic-sync.db atomic-sync.db.backup
cp atomic-sync.db-wal atomic-sync.db-wal.backup 2>/dev/null || true
cp atomic-sync.db-shm atomic-sync.db-shm.backup 2>/dev/null || true
```

An online SQLite backup tool is preferable when zero downtime is required. Never restore only the main file while an unmatched WAL remains.

The dedicated rclone config directory must be backed up without replacing a newer refreshed OAuth token with an older file during restore.

## Upgrade

1. Read release notes and database compatibility notes.
2. Back up SQLite and the dedicated rclone config directory.
3. Pull the new image digest without recreating containers.
4. Render the updated Compose model and confirm the root filesystem is read-only, only `/data` and `/config/rclone` are writable by default, and the media source still has the intended permission.
5. Recreate only `atomic-sync`; never run an unscoped `docker compose up -d` in a busy shared stack.
6. Wait for `/api/ready`, inspect startup logs, and confirm UI version/commit and recent run states.
7. Keep all jobs paused or dry-run until the upgraded process and database have been verified.

When upgrading from v0.1.2 or earlier, replace `RCLONE_CONFIG_PATH=<file>` with `RCLONE_CONFIG_DIR=<dedicated-directory>`, own the directory as UID/GID 1000 with mode `0700`, and set `rclone.conf` to `0600`.

### v0.1 legacy staging inventory

Before or immediately after upgrading from v0.1, independently inventory `.atomic-sync-staging` on every destination. Version 0.1 could retain verified staging as recovery evidence after an immutable merge. Version 0.2 does not create, publish, transfer, or delete that namespace. Job validation rejects a source or destination endpoint when any normalized path segment is exactly `.atomic-sync-staging`. A valid parent destination may still contain a legacy child with that name; destination branch analysis deliberately excludes the child. Source discovery also fails closed if it encounters the reserved namespace below an allowed source endpoint.

Leaving legacy staging in place is safe from v0.2 automation, but it still consumes storage. Record its paths and sizes, keep writers quiesced for any content comparison, and compare it independently with the authoritative source and final destination. Use a reviewed content check such as `rclone check --download` when deletion requires byte-level proof. Only an operator may remove legacy staging after the evidence and a recovery copy are no longer required; Atomic Sync never cleans it automatically.

## Rollback

If health, state, or logs regress:

1. Stop only Atomic Sync and prevent new runs.
2. Preserve current database, logs, and branch inventories.
3. Restore the previous image digest.
4. Preserve the newest dedicated rclone config; do not overwrite a refreshed token with an older backup.
5. Restore SQLite only when the release notes require it or the schema is incompatible.
6. Start only Atomic Sync, verify `/api/ready`, and leave jobs paused.
7. Reconcile any direct-transfer partial state before retrying or resuming writers.

An interrupted non-terminal run becomes `failed` on restart. This changes control-plane state only; it does not undo files already handled by rclone.

## Log review

Healthy startup contains one structured `atomic-sync listening` entry with version and commit. Investigate:

- `job run failed`, `unit failed`, or interrupted non-terminal runs;
- invalid `mode`/`deleteSource` pairing or missing move confirmation;
- shallow directory-unit or parent/child overlap errors;
- destination-exists or immutable conflict errors;
- verification/hash-unavailable errors;
- repeated timeouts or retries;
- Drive `403`/`429` quota errors;
- source permission errors during a real move;
- SQLite busy, read-only, or disk-full errors.

The application never logs the API token or rclone configuration. Rclone error output is bounded before it is attached to application logs.

## Drive quota pressure

When Drive reports quota errors:

1. Pause new Atomic Sync runs and analyses.
2. Do not repeatedly restart the rclone mount or retry destructive jobs.
3. Identify scanners, mounts, and other clients using the same OAuth project.
4. Let the provider window recover.
5. Resume with one metadata analysis or dry-run, then one transfer job at low concurrency.

Atomic Sync serializes branch analysis, but every application using the same Drive project still shares provider quota.
