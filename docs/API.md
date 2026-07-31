# HTTP API

All responses are JSON except the embedded UI and the SSE event stream. When `ATOMIC_API_TOKEN` is set, every `/api/*` route except health and readiness requires:

```http
Authorization: Bearer <token>
```

The UI loads without authentication and asks for the token before protected requests. This avoids placing a token in a query string or URL log.

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
| `GET` | `/api/jobs/{id}/analysis` | Yes | Get one job's latest analysis |
| `POST` | `/api/jobs/{id}/analysis` | Yes | Start a serialized read-only branch analysis |
| `GET` | `/api/runs?limit=100` | Yes | Recent durable run records; limit 1–500 |
| `GET` | `/api/events` | Yes | SSE run/analysis events; use a fetch stream to send the Bearer header |

Unknown JSON fields, multiple JSON documents, bodies over 1 MiB, and invalid enum/path values are rejected.

## v0.1.0 execution boundary

Version 0.1.0 is intentionally copy-only. Jobs must use `"mode": "copy"` and `"deleteSource": false`; create/update requests with `move` or source deletion enabled are rejected. The Runner repeats this validation when a stored job starts, so a legacy or manually altered database record cannot bypass the API boundary. Atomic Sync does not expose a source-cleanup endpoint.

Source cleanup is an external administrative operation. Stop every writer for the selected directory, perform an independent final verification, confirm recovery media, and delete only the reviewed unit outside Atomic Sync. See [Operations](OPERATIONS.md#manual-source-cleanup-outside-atomic-sync).

An active job cannot be updated or deleted. After the first unit receives a durable destination assignment, source, grouping/depth, destination names, paths, weights, and ordering are placement-locked; create a new job to change them safely. Other successful updates delete the old analysis snapshot so the UI cannot present results from a previous configuration.

The API returns `409 Conflict` when any source or destination is equal to, a
parent of, or nested below a path already owned by another job. This check is
also repeated before runs and analyses so legacy overlapping records fail
closed instead of executing concurrently. Job starts, analysis starts, updates,
deletes, and overlap checks share a short serialization boundary, so a worker
cannot begin with one configuration while the database is being replaced by
another.

Archive-analysis units include source/destination presence, file and byte
totals, matching/missing/conflicting paths, destination-only counts and bounded
evidence samples. Multi-destination jobs also report files and branch names
found outside the unit's assigned destination; those units fail closed as
`conflict`. These fields describe a metadata snapshot, not checksum proof.

## Minimal dry-run job

```json
{
  "name": "Archive stable movies",
  "source": "/sources/storagebox/movies",
  "destinations": [
    {"name": "gd-primary", "path": "GD:data/media/movies", "weight": 1}
  ],
  "mode": "copy",
  "grouping": "folder",
  "settleSeconds": 2592000,
  "concurrency": 2,
  "verify": "checksum",
  "conflictPolicy": "fail",
  "dryRun": true,
  "paused": false
}
```

Valid groupings are `folder`, `show`, `season`, and `depth`. Valid conflict policies are `fail` and `merge-immutable`.

Executable units must resolve to directories at one fixed hierarchy:

- `folder` and `show` use one top-level directory.
- `season` uses a two-level `Show/Season` directory.
- `depth` uses exactly the configured positive depth.

A media file above that boundary is shallow; a discovery containing both a parent and child unit is ambiguous. Either condition fails the entire run before staging. Branch analysis can still report the malformed physical structure so it can be repaired.

## Verification modes

`verify: checksum` invokes `rclone check --download`. Every compared file is read in full from both endpoints; this works without a common backend hash but can consume substantial source, network, and Drive bandwidth. `verify: size` uses size-only comparison and reads metadata rather than file contents, so it is weaker.

The source-to-staging check is bidirectional and exact within the job's filter set: extra, missing, or different selected staging objects fail the run. A new destination receives another bidirectional exact check. Only `merge-immutable` uses a one-way final completeness check, allowing reviewed destination-only posters, subtitles, and previously archived objects to remain.

`include` and `exclude` filters are rejected in v0.1.0. Every executable directory unit is copied and verified in full, preventing a filter from publishing an empty shell or silently splitting related media metadata from the unit. Move jobs are unsupported as well.

Local sources must be below `/sources`; local destinations must be below `/destinations`. Remote sources and backslashes in local paths are rejected. This prevents an API-token holder from selecting control-plane files such as `/config/rclone.conf` as media. A configured rclone remote may be used as a destination, so the API token remains a full administrative credential for every destination remote present in the mounted rclone configuration.
