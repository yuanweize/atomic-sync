# Changelog

All notable changes follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and semantic versioning.

## [Unreleased]

## [0.3.0] - 2026-08-02

### Added

- Progressive bilingual job editor with plain-language copy/move outcomes, human-readable stable-window units, hierarchy-only media presets, inline safety guidance, and a live source-to-destination execution summary.
- Complete multi-destination editing for names, paths, and routing weights, plus an automated UI contract check for translation parity, references, responsive rules, and required form behaviors.

### Changed

- Position Atomic Sync as a regular-file, directory-unit transfer control plane; `folder` and `depth` are general rules, while `show` and `season` remain hierarchy-only media presets over the same engine.
- Clarify in the UI and documentation that the API token is not a system or Tailscale password: only loopback listeners may omit it, while direct Tailscale-IP and reference-container listeners still require at least 32 characters.

### Fixed

- Stop hiding secondary destinations during edits; every saved destination and weight is now visible and round-trips through the editor.
- Show custom depth only when it applies, preserve legacy schedules explicitly instead of presenting a misleading disabled field, and retain unused legacy depth values on no-op edits so placement locks are not tripped accidentally.

## [0.2.0] - 2026-08-01

### Added

- Direct move mode for high-throughput source-to-destination archival through rclone.
- Destructive-intent pairing: `copy` requires `deleteSource: false`, while `move` requires `deleteSource: true`; conflict policy and verification remain independent operator choices.
- Exact job-name confirmation for non-dry-run move execution through both the API and web control plane.
- Real rclone dry-runs that exercise destination access and transfer planning without modifying source or destination media objects; OAuth token refresh remains possible.
- Per-unit stable-window revalidation immediately before every rclone operation.
- Per-invocation temporary `--files-from-raw` manifests generated from discovery fingerprints, pinning rclone to the reviewed file set without creating staging or a media copy.
- Optional uncapped rclone TPS mode (`ATOMIC_RCLONE_TPS_LIMIT=0`) for measured dedicated-OAuth deployments.

### Changed

- Use rclone as the sole transfer data plane and write each unit directly to the assigned final destination in one transfer.
- Stream checks and transfers without `--check-first` to reduce startup latency and improve throughput.
- Protect every non-dry-run copy or move with a final path-and-size inventory against the discovery fingerprint; move additionally uses native `--ignore-existing` and a source-residue check so a missing discovered object or destination overlap cannot be reported complete.
- Clarify that move overlap paths skipped by `--ignore-existing` are not checksum evidence and require an independent content check before cleanup.
- Update the minimal rclone build to 1.75.0, including upstream local-path and dependency security fixes.
- Reduce the public model to exactly two operations: source-preserving `copy` and source-removing `move`. Rclone `sync` is intentionally absent because it may delete destination-only content.
- Execute `merge-immutable` directly against the final destination with overwrite protection, preserving destination-only files.
- Map checksum verification to rclone's `--checksum` comparison so compatible backend hashes can be used without downloading both copies; keep `--size-only` as the fastest metadata path.
- Rework the bilingual documentation and control plane around complete-directory transfers, mergerfs branch evidence, explicit destructive intent, recovery, and throughput tuning.
- Keep the repository stable-window default at 30 days; document three days only as a scoped dry-run/canary profile.
- Pass a positive stable window to rclone as `--min-age <seconds>s`, while the pinned manifest prevents files arriving after revalidation from joining the transfer.
- Reserve `.atomic-sync-staging` at the endpoint boundary: source or destination endpoints containing that exact path segment are rejected, while destination analysis ignores a legacy child below an otherwise valid parent. Version 0.2 does not create, transfer, or delete the namespace, and source discovery fails closed when it is encountered.

### Security

- Keep the reference Compose source bind read-only and all new jobs dry-run by default; production move requires an explicit writable-source opt-in.
- Preserve immutable destination conflict handling and cross-job path-overlap rejection for both modes.
- Revalidate the discovery fingerprint immediately before rclone, pin its file set in a temporary manifest, and require every discovered path and size after every non-dry-run copy or move; move then checks source residue. These metadata gates narrow time-of-check/time-of-use races without a second content transfer, but are not writer locks or proof against preserved-mtime equal-size rewrites.
- Require a dedicated Google OAuth client before aggressive Drive concurrency tuning, and document the multiplication of process and per-process transfer limits.

## [0.1.3] - 2026-08-01

### Fixed

- Mount only the dedicated rclone config directory read-write so OAuth token refresh can persist via atomic rename while the container root and media source remain read-only.
- Replace the Compose interpolation variable `RCLONE_CONFIG_PATH` with `RCLONE_CONFIG_DIR`; deployments with a custom config location must migrate it to the dedicated directory path.
- Keep a stable login-form reference across asynchronous authentication, with disabled/loading feedback and recoverable focus on rejected tokens.

## [0.1.2] - 2026-07-31

### Fixed

- Allow only rclone's exact shared-Google-Drive-client retirement `NOTICE` during inventory scans; every other or mixed `lsjson` diagnostic remains fail-closed.

## [0.1.1] - 2026-07-31

### Fixed

- Accept the empty `ModTime` value emitted by `rclone lsjson --no-modtime` during branch analysis while still failing closed when a stable-window decision requires an unknown modification time.

## [0.1.0] - 2026-07-31

### Added

- Physical-branch archive analysis for mergerfs-backed media libraries.
- File-level archived, ready-to-verify, partial (including fully disjoint
  branches), pending, conflict, and empty classifications with bounded evidence.
- Bilingual, token-aware operational UI.
- Dry-run, copy-only, exact-staging, immutable-merge, shutdown, API, store, and branch-analysis tests.
- Signed multi-architecture GHCR release pipeline with SBOM and provenance.

### Security

- Copy-only v0.1.0 boundary: API and Runner reject move/`deleteSource`, the reference source is read-only, and the minimal rclone image omits `purge`.
- Directory-only execution units at a fixed hierarchy, with shallow and parent/child overlap discovery failing before staging.
- Bidirectional exact selected-source-to-staging verification and selected-source-complete final verification that preserves immutable destination extras.
- Explicit verification modes: checksum uses `rclone check --download` and reads full file contents from both endpoints; size mode compares metadata only.
- Successful `merge-immutable` staging is retained for recovery/audit; Atomic Sync performs no automatic staging or source cleanup.
- Server-generated job IDs, strict JSON decoding, CSP, security headers, and constant-time token checks.
- Non-root, read-only, capability-free container defaults.
- Cross-job source/destination overlap rejection and visible mobile-safe conflict evidence.
- Job mutation/start serialization and fail-closed multi-destination placement checks.
- Security-rebuilt, minimal rclone 1.74.4 runtime with fixed `x/text` and
  `grpc` dependencies instead of suppressing known HIGH findings.
- Release-time verification of the Cosign identity, SBOM, provenance, and
  anonymous GHCR access.

[Unreleased]: https://github.com/yuanweize/Atomic-Sync/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/yuanweize/Atomic-Sync/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/yuanweize/Atomic-Sync/compare/v0.1.3...v0.2.0
[0.1.3]: https://github.com/yuanweize/Atomic-Sync/releases/tag/v0.1.3
[0.1.2]: https://github.com/yuanweize/Atomic-Sync/releases/tag/v0.1.2
[0.1.1]: https://github.com/yuanweize/Atomic-Sync/releases/tag/v0.1.1
[0.1.0]: https://github.com/yuanweize/Atomic-Sync/releases/tag/v0.1.0
