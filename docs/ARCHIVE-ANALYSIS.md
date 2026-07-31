# Branch-aware archive analysis

Union filesystems such as mergerfs merge directory names and directory contents. A path visible at `/media/merged/Show` may contain files from both `/media/source/Show` and `/media/archive/Show`. The merged path is useful to consumers such as Sonarr, but it cannot answer whether the source branch has been fully archived.

Atomic Sync always analyzes the physical job source and its assigned physical destination. Never configure the mergerfs union as both the source of truth and the archive destination.

## Decision model

The analyzer runs `rclone lsjson --recursive` once for the source and once for each configured destination. It groups every entry using the job's atomic-unit policy, then compares each source file with the same relative path at the selected destination.

```text
source has no files + destination files present  → archived
source files present + destination unit absent   → pending
destination has files but some source files are missing → partial
all source paths/sizes match                     → ready-to-verify
same relative path, different size               → conflict
same relative path, file/directory type differs  → conflict
files exist outside the unit's assigned destination → conflict
no files on either branch                        → empty
```

“Source has no files” includes an empty source directory shell. This matters with mergerfs: the shell may remain visible on StorageBox even though every real object lives on GD. The API exposes `sourcePresent` and `destinationPresent` so operators can distinguish an absent unit from an empty shell without changing the macro status.

Destination-only extra files do not reduce source coverage. They may have been archived earlier or created by a media manager. They are never removed by analysis. A unit can therefore be `partial` with 0% source coverage when its two physical branches contain entirely complementary files: the destination already holds part of the merged unit, but none of the files still on the source has reached that destination yet.

Analysis intentionally describes the complete physical unit. Version 0.1.x rejects `include`/`exclude` filters, move mode, and `deleteSource`; analysis never authorizes or performs source cleanup.

Any source or destination listing error fails the whole analysis. An exhausted cloud API is never converted into an empty inventory or a false `archived` result.

A successful zero-file source listing is valid when the whole library has been archived. Filesystem availability must therefore be verified independently before analysis. The reference Compose model sets `create_host_path: false`, and operators should verify the expected mount type with `findmnt` or an equivalent health check before starting the container. No file inventory can distinguish a genuinely empty share from an existing but unmounted empty directory.

For jobs with multiple weighted destinations, existing SQLite assignments win. An unassigned source unit uses the same deterministic weighted selection as execution without persisting a new assignment. A destination-only unit is consolidated into one logical result on the branch that contains it. If files for that unit also exist on another configured destination, the result becomes `conflict`; empty directory shells on secondary destinations remain presence metadata and do not create a false conflict.

## What “100%” means

Dashboard coverage is the percentage of source file paths whose destination size matches. It is a low-I/O planning signal, not proof that file contents are identical. `archived` is displayed as 100% because no source files remain; this is completion state, not historical proof of how the destination files arrived.

Calculating source checksums across a multi-terabyte CIFS mount merely to refresh a dashboard would be expensive and could disrupt playback. Atomic Sync therefore reserves content verification for an explicit run:

1. Copy source to hidden staging.
2. Verify the selected source set and staging bidirectionally; they must match exactly under the same filters.
3. Publish or immutable-merge staging.
4. Verify the original source against the final destination: exact and bidirectional for a new destination, or one-way completeness for `merge-immutable` so reviewed destination-only objects can remain.
5. Mark the verified copy complete and retain the source.

`verify: checksum` currently runs `rclone check --download`. It reads every compared file in full from both endpoints, avoiding backend-hash compatibility assumptions at the cost of substantial source and cloud I/O. `verify: size` uses size-only comparison, reads metadata instead of file contents, and offers weaker assurance.

For a new destination, promotion moves the hidden destination-side staging directory to its final name. For `merge-immutable`, the hidden staging copy is retained after a successful run as recovery and audit material; Atomic Sync never performs automatic staging cleanup.

## Analysis units versus executable units

The analyzer records malformed or scattered physical layouts so operators can repair them, but execution is stricter. A runnable unit must resolve to a directory at one fixed grouping depth: one top-level directory for `folder`/`show`, `Show/Season` for `season`, or exactly the configured `depth`. A shallow media file or a discovered parent/child unit overlap fails the entire run before staging. This prevents a root-level episode and its season directory from being copied as two independent units.

An analysis status, including `ready-to-verify`, is not permission to delete source data. Source cleanup is outside Atomic Sync v0.1.x and requires the quiesced manual workflow in [Operations](OPERATIONS.md#manual-source-cleanup-outside-atomic-sync).

## Interpreting overlapping folders

Consider a show with two episodes:

```text
StorageBox/Show/Season 01/E01.mkv
StorageBox/Show/Season 01/E02.mkv
GD/Show/Season 01/E01.mkv
```

mergerfs displays both episodes inside one `Show` directory. Atomic Sync reports the unit as `partial`, with 1/2 matching source files and one missing sample. It does not call the show archived merely because both branches contain `Show`.

If the GD copy of `E01.mkv` has a different size, the status becomes `conflict`. `merge-immutable` refuses to overwrite it. Select the authoritative file manually before retrying.

The overlap can also be completely scattered:

```text
StorageBox/Movie/main.mkv
GD/Movie/poster.jpg
```

The merged directory contains both files, but GD contains none of the source-side payload. Atomic Sync reports `partial` with 0% source coverage, not `archived` and not `ready-to-verify`. If GD has only an empty `Movie` directory, the result is instead `pending`; an empty directory shell is not evidence that any file was archived.

If both source files match GD but remain on StorageBox, the status is `ready-to-verify`, not `archived`. This distinction prevents an unverified duplicate from being mistaken for successful cleanup.

## Performance and API quota

- Analysis is read-only and metadata-first.
- Only one analysis runs at a time per Atomic Sync process.
- Destinations are listed sequentially to avoid bursts against cloud APIs.
- Repeated manual refreshes should be avoided while a Drive API quota is exhausted.
- Execution concurrency and analysis concurrency are separate limits.

Analysis results are persisted in SQLite. The UI can show the latest result without rescanning the branches.

## Empty directories

Some object stores do not preserve empty directories, while CIFS does. `empty` is informational. Do not create or delete content solely to make empty-directory counts match.
