# Changelog

All notable changes follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and semantic versioning.

## [Unreleased]

### Added

- Physical-branch archive analysis for mergerfs-backed media libraries.
- Bilingual, token-aware operational UI.
- Dry-run, immutable-merge, shutdown, API, store, and branch-analysis tests.
- Signed multi-architecture GHCR release pipeline with SBOM and provenance.

### Security

- Fail-closed destination policy and final source-to-destination verification.
- Server-generated job IDs, strict JSON decoding, CSP, security headers, and constant-time token checks.
- Non-root, read-only, capability-free container defaults.
- Security-rebuilt, minimal rclone 1.74.4 runtime with fixed `x/text` and
  `grpc` dependencies instead of suppressing known HIGH findings.

[Unreleased]: https://github.com/yuanweize/atomic-sync/compare/v0.1.0...HEAD
