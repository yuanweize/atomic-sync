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

An active job cannot be updated or deleted. After the first unit receives a durable destination assignment, source, grouping/depth, destination names, paths, weights, and ordering are placement-locked; create a new job to change them safely. Other successful updates delete the old analysis snapshot so the UI cannot present results from a previous configuration.

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

`include` and `exclude` rclone filters are supported only in `copy` mode. Atomic Sync rejects filters in `move` mode because cleanup is intentionally unit-wide; accepting that combination could delete excluded files that were never copied.
