# HTTP API

All responses are JSON except the embedded UI and the SSE event stream. When `ATOMIC_API_TOKEN` is set, every `/api/*` route except health and readiness requires:

```http
Authorization: Bearer <token>
```

The UI loads without authentication and asks for the token before protected requests. It keeps the token in the current tab's `sessionStorage`, never in a query string or URL log.

## Endpoints

| Method | Path | Auth | Description |
|---|---|---:|---|
| `GET` | `/api/health` | No | Process liveness, version, and whether auth is required |
| `GET` | `/api/ready` | No | SQLite readiness |
| `GET` | `/api/system` | Yes | Version, commit, build time, and active job count |
| `GET` | `/api/dashboard` | Yes | Job/run counters |
| `GET` | `/api/jobs` | Yes | List jobs |
| `POST` | `/api/jobs` | Yes | Create a job; server generates the ID; omitted `dryRun` becomes `true` |
| `GET` | `/api/jobs/{id}` | Yes | Read a job |
| `PUT` | `/api/jobs/{id}` | Yes | Replace an idle job, preserve identity, and invalidate its old analysis |
| `DELETE` | `/api/jobs/{id}` | Yes | Delete an idle job and its assignments/latest analysis |
| `POST` | `/api/jobs/{id}/run` | Yes | Start a bounded asynchronous run |
| `GET` | `/api/analyses` | Yes | List latest analysis summaries; unit details are omitted |
| `GET` | `/api/jobs/{id}/analysis` | Yes | Get one job's latest physical-branch analysis |
| `POST` | `/api/jobs/{id}/analysis` | Yes | Start a serialized read-only branch analysis |
| `GET` | `/api/runs?limit=100` | Yes | Recent durable run records; limit 1–500 |
| `GET` | `/api/events` | Yes | SSE run/analysis events; use a fetch stream to send the Bearer header |

Unknown JSON fields, multiple JSON documents, bodies over 1 MiB, and invalid enum/path values are rejected.

## Two execution modes

Jobs must use exactly one of these mode/intent pairs:

| `mode` | `deleteSource` | Operation | Source after a successful transfer |
|---|---:|---|---|
| `copy` | `false` | Direct rclone `copy` to the final destination | Preserved |
| `move` | `true` | Direct rclone `move` to the final destination | Removed by rclone after successful transfer |

`sync` is rejected. Rclone `sync` can delete destination-only objects and is intentionally not part of Atomic Sync's API.

Conflict policy and verification are independent of mode. Both copy and move accept `fail` or `merge-immutable`, and `size` or `checksum`. This allows a fast `move + fail + size` dry-run canary while still supporting a stronger checksum-based immutable retry after a reviewed partial transfer.

The Runner validates the contract again when a stored job starts, so a manually altered or legacy SQLite record cannot bypass the API boundary. Both modes invoke rclone directly against the final destination; rclone owns transfer retries, resumability, and transfer comparison:

- `verify: checksum` maps to `--checksum`;
- `verify: size` maps to `--size-only`.

Every invocation also receives a temporary `--files-from-raw` manifest generated from the discovery fingerprint. It pins the transfer to the reviewed file set, is deleted after the invocation, and contains paths only—not staging data or media bytes. When `settleSeconds > 0`, the Runner also passes `--min-age <settleSeconds>s`.

The run states are `discovered` → `transferring` → `completed` or `failed`.

For non-dry-run `move`, the request must include the exact job name in:

```http
X-Atomic-Confirm-Job: <exact job name>
```

The comparison is constant-time. Dry-run move requests do not require the header because no source or destination media object is written or removed; rclone may still refresh OAuth credentials in its dedicated config directory.

## Minimal dry-run job

```json
{
  "name": "Archive stable movies",
  "source": "/sources/media/movies",
  "destinations": [
    {"name": "gd-primary", "path": "GD:data/media/movies", "weight": 1}
  ],
  "mode": "copy",
  "deleteSource": false,
  "grouping": "folder",
  "settleSeconds": 2592000,
  "concurrency": 1,
  "verify": "size",
  "conflictPolicy": "fail",
  "dryRun": true,
  "paused": true
}
```

The API, reference UI, and example configuration default an omitted stable window to 30 days (`2592000` seconds). An explicit zero remains zero. Use `259200` only for a deliberately scoped three-day dry-run/canary.

Valid groupings are `folder`, `show`, `season`, and `depth`. Valid conflict policies are `fail` and `merge-immutable`.

Executable units must resolve to directories at one fixed hierarchy:

- `folder` and `show` use one top-level directory;
- `season` uses a two-level `Show/Season` directory;
- `depth` uses exactly the configured positive depth.

A media file above that boundary is shallow; a discovery containing both a parent and child unit is ambiguous. Either condition fails the entire run before rclone writes. Branch analysis can still report the malformed physical structure so it can be repaired.

## Conflict policies

`fail` stops when the destination unit already exists. It is the safest first canary policy for either copy or move because no existing unit is merged implicitly.

`merge-immutable` writes only missing files to an existing final directory and never overwrites a destination object. For move, native `--ignore-existing` also keeps every overlapping source path. After every non-dry-run copy or move returns, Atomic Sync lists the final destination and requires every discovered file path and size to be present; move then checks for remaining source files. Either failure records the unit as failed, with a move treated as partial. Ignored paths are not compared by `verify: checksum` or `verify: size`, so they still require independent content proof. This avoids deleting a same-size conflict or a checksum comparison that fell back to size. A direct merge can still leave successfully added files at the destination, so review branch analysis before retrying.

After a partial or interrupted operation, review the remaining source and destination inventories before changing the task to `merge-immutable` for a retry. Neither policy uses `sync` semantics or prunes destination-only files.

## Verification semantics

`verify: checksum` enables rclone's `--checksum` comparison for paths rclone considers during the transfer. For a local/CIFS source and Google Drive, compatible MD5 hashes are normally compared without downloading the Drive object. `verify: size` enables `--size-only` and is the fastest metadata path. On move, `--ignore-existing` takes precedence for a destination-overlap path, so neither verification mode proves that retained source object's content. Hash availability and consistency are backend properties.

Immediately before every operation, Atomic Sync re-lists the source and requires its paths, types, sizes, and modification times to match the discovery fingerprint while stable-window eligibility remains valid. Rclone receives only those file paths through `--files-from-raw`; a positive stable window is also enforced by `--min-age`. After every non-dry-run copy or move, Atomic Sync runs `lsjson` against the final destination and requires every discovered file at the same path and size; `fail` also rejects unexpected final paths, and move then checks source residue. The manifest can reduce redundant source traversal and the final listing transfers no media content, but this metadata closure is not a second content verification, checksum, or `rclone check`. Writers must still be quiesced because an in-place, equal-size rewrite that preserves its modification time cannot be proven from the fingerprint.

An operator may independently run `rclone check --download` for a selected, quiesced deep audit. Atomic Sync does not invoke this slower full-content check in the normal run path.

## Placement and analysis rules

`job ID + unit path` is the assignment key. The first weighted destination selection is persisted in SQLite and reused across retries. After the first assignment, source, grouping/depth, destination names, paths, weights, and ordering are placement-locked; create a new job to change them safely.

The API returns `409 Conflict` when any source or destination is equal to, a parent of, or nested below a path already owned by another job. This check is repeated before runs and analyses so legacy overlapping records fail closed instead of executing concurrently. Job starts, analysis starts, updates, deletes, and overlap checks share a short serialization boundary.

Archive-analysis units include source/destination presence, file and byte totals, matching/missing/conflicting paths, destination-only counts, and bounded evidence samples. A metadata snapshot does not prove content identity or authorize source removal. See [Branch-aware archive analysis](ARCHIVE-ANALYSIS.md).

## Endpoint restrictions

Local sources must be below `/sources`; local destinations must be below `/destinations`. Remote sources and backslashes in local paths are rejected. The legacy `.atomic-sync-staging` namespace is reserved: validation rejects any source or destination endpoint whose normalized path contains a segment with that exact name. A valid parent destination may contain a legacy child with that name, and destination analysis ignores the child; source discovery fails closed if it encounters one below an allowed source endpoint. This prevents an API-token holder from selecting control-plane files such as `/config/rclone.conf` as media. A configured rclone remote may be used as a destination, so the API token remains administrative access to every destination remote in the mounted configuration.
