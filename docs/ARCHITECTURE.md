# Architecture

Atomic Sync is a single Go binary with an embedded browser UI, a REST/SSE API, a bounded rclone execution engine, and a single-writer SQLite store.

## Components

```mermaid
flowchart TB
  UI[Embedded bilingual SPA] -->|Bearer API| API[net/http API]
  API --> STORE[(SQLite WAL)]
  API --> RUNNER[Runner lifecycle]
  RUNNER --> STORE
  RUNNER -->|argv, no shell| RCLONE[rclone]
  RCLONE --> SOURCE[Physical source branch]
  RCLONE --> STAGE[Hidden destination staging]
  RCLONE --> FINAL[Final destination branch]
  RUNNER -->|SSE events| UI
```

- `cmd/atomic-sync`: configuration, logging, signal handling, recovery, and HTTP lifecycle.
- `internal/api`: authenticated JSON endpoints, static UI, CSP/security headers, and SSE.
- `internal/engine`: grouping, deterministic destination selection, analysis, bounded execution, and the publication protocol.
- `internal/model`: job validation and state rules.
- `internal/store`: SQLite schema, jobs, assignments, analyses, and run history.
- `internal/api/ui`: dependency-free HTML/CSS/JavaScript UI.

## Run state machine

```mermaid
stateDiagram-v2
  [*] --> discovered
  discovered --> completed: dry run
  discovered --> staging
  staging --> verifying
  verifying --> publishing
  publishing --> completed
  discovered --> failed
  staging --> failed
  verifying --> failed
  publishing --> failed
```

`failed` is terminal for one run record. A retry creates a new run and preserves the previous evidence and staging path.

## v0.1.0 safety boundary

Version 0.1.0 is copy-only. `mode: move` and `deleteSource: true` are rejected by both model/API validation and the Runner's execution-boundary validation. The official minimal rclone build omits `purge`, so the process has no source-deletion command to invoke. Source cleanup is a separate, quiesced operator workflow documented in [Operations](OPERATIONS.md#manual-source-cleanup-outside-atomic-sync).

The NUE first rollout mounts the physical StorageBox source read-only and starts with separate movie and TV dry-run jobs. Enabling a destination write canary does not change the source mount to read-write.

## Execution-unit contract

An executable unit is always a directory at one fixed hierarchy boundary:

- `folder` and `show`: one top-level directory.
- `season`: exactly `Show/Season`.
- `depth`: exactly the configured positive number of directory components.

A root-level media file, a show-level file in a season job, or any other item above the configured boundary is shallow and cannot become an executable file unit. Discovery also rejects a set containing both a directory and one of its descendants as separate units. These checks fail the complete run before staging, preventing mixed layouts such as `Show` plus `Show/Season 03` from being published independently.

## Destination assignment

`job ID + unit path` is the assignment key. A deterministic FNV-1a weighted selection chooses an initial destination, then SQLite persists it with `INSERT OR IGNORE`. Concurrent workers therefore converge on the same destination.

After the first assignment, the API placement-locks the source, grouping boundary, destination names, paths, weights, and ordering. This prevents an edit from silently splitting an already assigned media unit across physical targets. Create a new job for a new placement policy.

## Execution concurrency

Each job uses a bounded worker pool based on `job.concurrency`. Workers also acquire a global semaphore configured by `ATOMIC_MAX_CONCURRENCY`, so multiple jobs cannot exceed the process-wide rclone-process ceiling. Every invocation additionally receives `ATOMIC_RCLONE_TRANSFERS`, `ATOMIC_RCLONE_CHECKERS`, and `ATOMIC_RCLONE_TPS_LIMIT` limits; process concurrency therefore does not silently multiply rclone's default internal parallelism. The engine never creates one goroutine per discovered unit.

Branch analyses use a separate single-slot semaphore and list destinations sequentially to reduce cloud API pressure.

## Shutdown and recovery

API-launched work belongs to the Runner context. On SIGTERM:

1. HTTP request contexts are canceled.
2. The server stops accepting work.
3. The Runner cancels rclone children and waits for workers.
4. SQLite closes after workers finish.

On the next start, any non-terminal run record or running analysis is marked `failed` with an interrupted message. Hidden staging data is never automatically deleted.

## Storage model

SQLite uses WAL mode, a five-second busy timeout, and one open connection. This deliberately supports one process and simplifies assignment consistency. Horizontal replicas are not supported.

Tables:

- `jobs`: validated JSON job documents.
- `runs`: append-only execution history plus state transitions.
- `assignments`: stable unit-to-destination mapping.
- `analyses`: latest persisted physical-branch comparison per job.
- `settings`: reserved for forward-compatible application settings.

## Publication semantics

The complete source directory unit is first copied to a hidden destination-side staging directory. Staging verification is bidirectional and exact: missing, extra, or different objects block publication.

For a new destination, verified staging is moved to the final path with `--immutable`; promotion consumes that destination-side staging path. For `merge-immutable`, missing staging files are copied to an existing final directory with `--immutable`, conflicts fail without overwriting, and the hidden staging copy is retained after success for recovery and audit. Atomic Sync does not automatically clean retained staging.

A new destination ends with an exact bidirectional source-to-final check. `merge-immutable` ends with a one-way completeness check so reviewed destination-only content can remain. `verify: checksum` adds `--download`, reading the full contents of every compared file from both endpoints. `verify: size` adds `--size-only` and compares paths and byte counts without reading file contents.

No successful or failed v0.1.0 transition deletes the source. The run reaches `completed` only after final verification, but completion means “verified copy published,” not “source removed.”

Remote object stores do not provide an ACID directory rename. The protocol guarantees preservation and verification boundaries, not a distributed filesystem transaction.
