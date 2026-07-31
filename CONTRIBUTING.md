# Contributing to Atomic Sync

Thank you for helping make media archival safer. Changes to transfer, verification, path handling, authentication, or persistence are safety-critical and require tests for both the success and failure paths.

## Development setup

Requirements:

- Go 1.25 or newer
- Docker Engine and Compose v2 for image tests
- rclone for optional local integration testing

```bash
git clone https://github.com/yuanweize/atomic-sync.git
cd atomic-sync
go test ./...
```

Never use production credentials or media in tests. `rclone.conf`, `.env`, SQLite state, local sources, and build output are ignored by Git and Docker, but contributors remain responsible for checking every staged file.

## Required checks

Run before opening a pull request:

```bash
gofmt -w .
go vet ./...
go test -race -count=1 -cover ./...
node --check internal/api/ui/app.js
ATOMIC_API_TOKEN="$(openssl rand -hex 32)" docker compose config --quiet
docker build -t atomic-sync:dev .
```

CI enforces a repository-wide coverage floor and scans the built image for fixable high/critical vulnerabilities.

## Test expectations

- Engine changes: fake-rclone command sequence, source-preservation failure path, and final-verification behavior.
- Analysis changes: source-only, destination-only, partial, conflict, ready-to-verify, and empty cases.
- API changes: authenticated and unauthenticated behavior, size limits, invalid input, and status codes.
- Store changes: migration compatibility, create/update/read/delete, and restart recovery.
- UI changes: English and Simplified Chinese labels, keyboard access, narrow viewport behavior, and no inline script/style that weakens CSP.

No test may delete or rewrite a real media path.

## Pull requests

Keep changes focused. Include:

1. User-visible behavior and motivation.
2. Safety invariants affected.
3. Failure and rollback behavior.
4. Tests run and their results.
5. Screenshots for UI changes when a browser is available.

Use Conventional Commit-style titles where practical, such as `feat:`, `fix:`, `docs:`, `test:`, `build:`, or `security:`.

## Releases

Maintainers create an annotated semantic-version tag after CI passes. The Release workflow re-verifies source, publishes a multi-architecture GHCR image, generates SBOM/provenance, signs the digest with Cosign, and creates checksummed Linux binaries and GitHub release notes.
