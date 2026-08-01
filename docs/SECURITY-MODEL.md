# Security model

Atomic Sync is an administrative control plane around rclone. Its security boundary covers job validation, directory grouping, path isolation, execution authorization, and auditable state. Rclone and the configured storage providers remain responsible for the actual byte transfer and provider-specific integrity behavior.

## Protected assets

- Source media, including files eligible for removal in real `move` mode.
- Destination media, including destination-only files that Atomic Sync must never prune.
- Rclone credentials, Google OAuth tokens, and remote definitions.
- SQLite jobs, destination assignments, analyses, and run history.
- The API token and operational path information returned by protected endpoints.

## Trust boundaries

The browser, reverse proxy, Atomic Sync HTTP server, Runner, rclone child process, local/CIFS source, cloud destination, SQLite store, and container runtime are separate trust zones.

Atomic Sync assumes that the host administrator, container runtime, mounted source, and dedicated `rclone.conf` are trusted. Anyone who has the API token can create jobs targeting any allowed destination remote present in that configuration; treat the token as an administrative secret.

## Destructive-operation boundary

Only two job modes exist:

- `copy` requires `deleteSource: false` and preserves source data;
- `move` requires `deleteSource: true` and delegates transfer plus successful-source removal to rclone.

Rclone `sync` is not exposed. Neither mode is allowed to prune destination-only content.

The reference Compose deployment combines several independent controls:

1. New jobs default to `dryRun: true`.
2. The source bind is read-only.
3. A real move requires the exact job name as an explicit API/UI confirmation.
4. Mode and `deleteSource` must form a valid pair, while conflict policy and verification must each be valid, at both model and Runner boundaries.
5. Operators must deliberately make only the required source bind writable before a move can delete anything.

These controls reduce accidental execution; they do not make an authorized move reversible. Rclone operates on individual objects, not as a cross-provider ACID directory transaction. It removes a source object only after rclone considers that object's transfer successful, but interruption can still leave the complete directory split across branches.

## Main threats and controls

| Threat | Control |
|---|---|
| Unauthenticated write or move | The reference/production deployment authenticates every protected API; only loopback development may omit a token, while non-loopback listeners require at least 32 characters |
| Accidental real move | Dry-run default, read-only reference source, exact job-name confirmation, and strict move/delete-source pairing |
| Destination-only content deleted | No public `sync` mode or destination-pruning path; both operations use copy/move semantics |
| Source deleted before rclone accepts its transfer | Source removal is delegated to rclone move's successful-transfer semantics rather than a separate application delete loop |
| Existing destination mistaken for the source object | Every move uses native `--ignore-existing`; overlaps remain at source and post-move residue makes the unit failed/partial |
| Ignored overlap mistaken for checksum proof | `--ignore-existing` takes precedence on move overlap paths; neither verification mode proves those objects, so cleanup requires an independent content check |
| Interrupted transfer looks complete in mergerfs | Every non-dry-run copy or move must pass a final path-and-size inventory against the discovery fingerprint; move then checks source residue. Failures remain durable evidence, and same-named merged directories never prove completion |
| Destination overwrite | Fail-closed existing-unit policy or explicit immutable merge; different existing objects are not overwritten |
| Partial immutable merge | Direct immutable merge can add missing files before a later conflict fails; source files not successfully moved remain available, and the partial destination is visible to analysis |
| Shallow or overlapping execution units | Only directories at one fixed grouping depth execute; shallow files and parent/child unit sets fail before rclone writes |
| Late writer expands or changes a unit | Immediate fingerprint revalidation, a pinned `--files-from-raw` set, positive-window `--min-age`, and every real transfer's final path-and-size gate narrow the race; production move still requires quiesced writers because preserved-mtime equal-size rewrites are not metadata proof |
| Timing comparison of API token | Constant-time byte comparison |
| Stored XSS through job data | Server-generated constrained IDs, DOM `textContent`, CSP, and no inline script |
| Command injection | `exec.CommandContext` argument vector; no shell interpolation; validated endpoints and rejected filters |
| Path traversal from listings | Relative path normalization and traversal rejection before joining |
| Control-plane files selected as media | Local sources restricted below `/sources`, local destinations below `/destinations`, and remote sources rejected |
| Source/destination self-overlap | Equal or nested endpoints on the same backend are rejected |
| Two jobs race on one tree | API serialization plus equal/nested cross-job path rejection before mutations, runs, and analyses |
| Unit silently changes destination | Deterministic assignment persisted in SQLite; placement-defining fields lock after first assignment |
| Crash leaves a misleading running record | Runner cancellation plus startup reconciliation of non-terminal records to `failed` |
| Credential copied into image or repository | `.gitignore`/`.dockerignore`, dedicated runtime config bind, and no API for reading rclone configuration |
| OAuth refresh broadens host write access | Only a dedicated private rclone config directory is writable; the container root remains read-only |
| Container breakout | Non-root UID, read-only root, zero capabilities, `no-new-privileges`, and process/resource limits |
| Dependency or image drift | Pinned bases/actions, Dependabot, CodeQL, race tests, vulnerability scanning, SBOM, provenance, and keyless signing |
| Unused rclone surface | Official image links only the required local/Drive/crypt backends and the commands exercised by the Runner |
| Cloud API exhaustion | Serialized analysis, sequential destination listing, bounded workers/transfers/checkers/TPS, and explicit operational backoff |

## Verification boundary

`verify: size` compares paths and byte counts. It avoids reading content and is useful for a fast canary, but cannot detect equal-size corruption.

`verify: checksum` maps to rclone's `--checksum` transfer comparison. Rclone uses a hash common to both backends when available. A local or CIFS-mounted source and Google Drive normally share MD5, allowing comparison with Drive's stored hash without downloading the destination object. With no common hash, rclone can fall back to size; this is not content-integrity proof. On move, `--ignore-existing` skips destination-overlap paths before either checksum or size comparison, so retained overlaps always require an independent content check. Rclone owns transfer verification, retries, and resumability.

Immediately before rclone, Atomic Sync revalidates the complete source fingerprint and stable window, writes the discovered file paths to a temporary `--files-from-raw` manifest, and passes `--min-age <seconds>s` when the window is positive. The manifest fixes rclone's transfer set and is deleted after use; it is neither staging nor a media copy. After every non-dry-run copy or move, Atomic Sync uses `lsjson` to require every discovered file path and size at the final destination; move then checks source residue. This detects a missing or resized discovered object and excludes files that arrive after revalidation, without a second content transfer. It remains metadata evidence—not checksum proof or `rclone check`—and does not remove the need to quiesce writers.

An independent `rclone check --download` reads both sides and can be used for a selected deep audit. It is not the normal Atomic Sync verification path and can consume substantial source, network, and Drive bandwidth.

Metadata branch analysis compares paths, types, and sizes only. Its `archived` and `ready-to-verify` statuses are planning evidence, not checksum guarantees or authorization to remove data.

## Deliberate limitations

- Direct copy/move favors throughput with one final-path transfer. There is no whole-directory cross-provider transaction or automatic rollback.
- A stable window, pinned manifest, `--min-age`, and pre/post metadata gates are not writer locks. Atomic Sync cannot prove that every importer, repair task, or manual uploader has stopped, nor detect an in-place equal-size content rewrite that preserves its old modification time.
- Rclone and each provider define retry, resume, hash, quota, and consistency semantics.
- Direct `merge-immutable` is not atomic and can leave successfully added files at the destination if a later conflict stops the unit.
- Move never auto-cleans an overlapping source path, even when it may be identical; independent proof and explicit operator cleanup are required.
- Any configured source or destination endpoint with a normalized path segment exactly named `.atomic-sync-staging` is rejected. A valid parent destination may contain a legacy child from v0.1, which destination analysis excludes; source discovery fails closed if it encounters such a child. Version 0.2 does not create, transfer, or delete it, and only an explicit, independently verified operator cleanup may remove it.
- The application token is a shared administrator secret, not user-level RBAC.
- SQLite supports one Atomic Sync process; horizontal replicas are not supported.
- The official image supports local, Google Drive, and crypt. Other providers require a reviewed custom build.
- The Docker daemon, host root, mount availability, and operator actions remain outside the application's isolation guarantee.
- No inventory can distinguish a healthy empty source from an existing mountpoint whose backing filesystem is offline; operators must check mounts independently.

Run the service on a private network or behind an identity-aware proxy even when Bearer authentication is enabled.
