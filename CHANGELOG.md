# Changelog

All notable changes follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and semantic versioning.

## [Unreleased]

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

[Unreleased]: https://github.com/yuanweize/atomic-sync/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/yuanweize/atomic-sync/releases/tag/v0.1.0
