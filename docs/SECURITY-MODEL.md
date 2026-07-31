# Security model

## Protected assets

- Source media, which the reference v0.1.x deployment mounts read-only and Atomic Sync never deletes.
- Destination media, which immutable publication must not overwrite.
- Hidden destination staging, including successful `merge-immutable` staging retained for recovery and audit.
- rclone credentials and cloud tokens.
- SQLite job/assignment/audit state.
- The API token and operational path information.

## Trust boundaries

The HTTP client, Atomic Sync process, rclone child process, local/CIFS source, cloud destination, container runtime, and reverse proxy are separate trust zones. Atomic Sync assumes the host administrator and mounted rclone configuration are trusted.

## Main threats and controls

| Threat | Control |
|---|---|
| Unauthenticated source deletion | v0.1.x rejects move/`deleteSource`, exposes no cleanup endpoint, and the official image omits rclone `purge`; Bearer auth still protects all write APIs |
| Timing comparison of token | Constant-time byte comparison |
| Weak or missing remote-access token | Tokens shorter than 32 characters are rejected; an empty token is allowed only on a loopback listener |
| Stored XSS through job IDs/names | Server-generated constrained IDs; DOM `textContent`; CSP with no inline script |
| Command injection | `exec.CommandContext` argument vector; no shell interpolation; validated endpoints/filters |
| Path traversal from a remote listing | Relative path normalization and traversal rejection before joining |
| Source/destination self-overlap | Validation rejects equal or nested endpoints on the same backend |
| Two jobs race on one tree | API serializes job mutations and starts, then rejects equal or nested paths across jobs |
| Destination overwrite | Fail-closed default and rclone `--immutable` publication |
| Partial publish followed by source deletion | Atomic Sync never deletes source data; the reference source bind is read-only |
| Shallow or overlapping execution units | Execution requires directory units at one fixed hierarchy; shallow and parent/child unit sets fail before staging |
| Staging drift or contamination | Source-to-staging verification is bidirectional and exact for the complete directory unit |
| Existing immutable destination extras | Only `merge-immutable` uses one-way final verification, so reviewed destination-only files are preserved rather than deleted; new destinations must match exactly |
| Crash leaves misleading running state | Lifecycle cancellation plus startup reconciliation to failed |
| Credential committed or copied into image | `.gitignore` and `.dockerignore` exclude rclone config and `.env` |
| Container breakout | Non-root UID, read-only root, zero capabilities, no-new-privileges, pids/memory/CPU limits |
| Dependency or base-image drift | Base-image digests, action SHAs, Dependabot, CodeQL, race tests, Trivy, SBOM/provenance/signing |
| Unused rclone backend attack surface | Official image links only local/Drive/crypt and the commands used by the engine; known fixed dependencies are rebuilt rather than ignored |
| Cloud API exhaustion | Serialized analysis, sequential destinations, bounded worker pools |

## Deliberate limitations

- The application token is a single shared secret, not user-level RBAC.
- Source cleanup is an external privileged operation. Atomic Sync cannot enforce that an operator has stopped every writer or eliminate a verification-to-manual-deletion race outside the container.
- rclone and the selected storage providers define hash availability and remote consistency.
- `verify: checksum` uses `rclone check --download` and reads the full contents of every compared file from both endpoints; it can materially affect CIFS, network, and cloud bandwidth. `verify: size` avoids content reads but is weaker.
- Successful `merge-immutable` runs retain hidden staging. Operators must treat it as protected media and remove it only through a reviewed external process.
- The official image supports local, Google Drive, and crypt rclone backends;
  other providers require a separately reviewed custom build.
- The Docker daemon and host root remain outside the application's isolation guarantee.
- Metadata analysis uses path and size; it is not a checksum guarantee.

Run the service on a private network or behind an identity-aware proxy even when the token is enabled.
