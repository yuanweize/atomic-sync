# Security model

## Protected assets

- Source media, which move mode can delete only after final verification.
- Destination media, which immutable publication must not overwrite.
- rclone credentials and cloud tokens.
- SQLite job/assignment/audit state.
- The API token and operational path information.

## Trust boundaries

The HTTP client, Atomic Sync process, rclone child process, local/CIFS source, cloud destination, container runtime, and reverse proxy are separate trust zones. Atomic Sync assumes the host administrator and mounted rclone configuration are trusted.

## Main threats and controls

| Threat | Control |
|---|---|
| Unauthenticated purge request | Bearer auth on protected API; reference Compose requires a token |
| Timing comparison of token | Constant-time byte comparison |
| Weak or missing remote-access token | Tokens shorter than 32 characters are rejected; an empty token is allowed only on a loopback listener |
| Stored XSS through job IDs/names | Server-generated constrained IDs; DOM `textContent`; CSP with no inline script |
| Command injection | `exec.CommandContext` argument vector; no shell interpolation; validated endpoints/filters |
| Path traversal from a remote listing | Relative path normalization and traversal rejection before joining |
| Source/destination self-overlap | Validation rejects equal or nested endpoints on the same backend |
| Destination overwrite | Fail-closed default and rclone `--immutable` publication |
| Partial publish followed by source deletion | Source-to-final verification is mandatory before purge |
| Crash leaves misleading running state | Lifecycle cancellation plus startup reconciliation to failed |
| Credential committed or copied into image | `.gitignore` and `.dockerignore` exclude rclone config and `.env` |
| Container breakout | Non-root UID, read-only root, zero capabilities, no-new-privileges, pids/memory/CPU limits |
| Dependency or base-image drift | Base-image digests, action SHAs, Dependabot, CodeQL, race tests, Trivy, SBOM/provenance/signing |
| Unused rclone backend attack surface | Official image links only local/Drive/crypt and the commands used by the engine; known fixed dependencies are rebuilt rather than ignored |
| Cloud API exhaustion | Serialized analysis, sequential destinations, bounded worker pools |

## Deliberate limitations

- The application token is a single shared secret, not user-level RBAC.
- A trusted operator can still configure a destructive move job.
- rclone and the selected storage providers define hash availability and remote consistency.
- The official image supports local, Google Drive, and crypt rclone backends;
  other providers require a separately reviewed custom build.
- The Docker daemon and host root remain outside the application's isolation guarantee.
- Metadata analysis uses path and size; it is not a checksum guarantee.

Run the service on a private network or behind an identity-aware proxy even when the token is enabled.
