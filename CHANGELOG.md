# Changelog

All notable changes follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and semantic versioning.

## [Unreleased]

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

[Unreleased]: https://github.com/yuanweize/atomic-sync/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/yuanweize/atomic-sync/releases/tag/v0.1.3
[0.1.2]: https://github.com/yuanweize/atomic-sync/releases/tag/v0.1.2
[0.1.1]: https://github.com/yuanweize/atomic-sync/releases/tag/v0.1.1
[0.1.0]: https://github.com/yuanweize/atomic-sync/releases/tag/v0.1.0
