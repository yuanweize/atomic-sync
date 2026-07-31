# Branch-aware archive analysis

Union filesystems such as mergerfs merge directory names and directory contents. A path visible at `/media/merged/Show` may contain files from both `/media/source/Show` and `/media/archive/Show`. The merged path is useful to consumers such as Sonarr, but it cannot answer whether the source branch has been fully archived.

Atomic Sync always analyzes the physical job source and its assigned physical destination. Never configure the mergerfs union as both the source of truth and the archive destination.

## Decision model

The analyzer runs `rclone lsjson --recursive` once for the source and once for each configured destination. It groups every entry using the job's atomic-unit policy, then compares each source file with the same relative path at the selected destination.

```text
source has no files + destination files present  → archived
source files present + destination unit absent   → pending
some source paths/sizes match, some are missing  → partial
all source paths/sizes match                     → ready-to-verify
same relative path, different size               → conflict
same relative path, file/directory type differs  → conflict
no files on either branch                        → empty
```

“Source has no files” includes an empty source directory shell. This matters with mergerfs: the shell may remain visible on StorageBox even though every real object lives on GD. The API exposes `sourcePresent` and `destinationPresent` so operators can distinguish an absent unit from an empty shell without changing the macro status.

Destination-only extra files do not reduce source coverage. They may have been archived earlier or created by a media manager. They are never removed by analysis.

Analysis intentionally describes the complete physical unit and therefore does not apply a copy job's `include` or `exclude` filters. Destructive move jobs reject filters entirely, preventing unit-wide cleanup from deleting excluded files.

Any source or destination listing error fails the whole analysis. An exhausted cloud API is never converted into an empty inventory or a false `archived` result.

A successful zero-file source listing is valid when the whole library has been archived. Filesystem availability must therefore be verified independently before analysis. The reference Compose model sets `create_host_path: false`, and operators should verify the expected mount type with `findmnt` or an equivalent health check before starting the container. No file inventory can distinguish a genuinely empty share from an existing but unmounted empty directory.

For jobs with multiple weighted destinations, existing SQLite assignments win. An unassigned source unit uses the same deterministic weighted selection as execution without persisting a new assignment. Destination-only units are reported for the branch on which they exist.

## What “100%” means

Dashboard coverage is the percentage of source file paths whose destination size matches. It is a low-I/O planning signal, not proof that file contents are identical. `archived` is displayed as 100% because no source files remain; this is completion state, not historical proof of how the destination files arrived.

Calculating source checksums across a multi-terabyte CIFS mount merely to refresh a dashboard would be expensive and could disrupt playback. Atomic Sync therefore reserves checksum work for an explicit run:

1. Copy source to hidden staging.
2. Run `rclone check` from source to staging.
3. Publish or immutable-merge staging.
4. Run `rclone check` from source to the final destination.
5. Only after step 4 may move mode purge the source unit.

Use `verify: size` only for backends without a useful common hash, and understand that it offers weaker assurance.

## Interpreting overlapping folders

Consider a show with two episodes:

```text
StorageBox/Show/Season 01/E01.mkv
StorageBox/Show/Season 01/E02.mkv
GD/Show/Season 01/E01.mkv
```

mergerfs displays both episodes inside one `Show` directory. Atomic Sync reports the unit as `partial`, with 1/2 matching source files and one missing sample. It does not call the show archived merely because both branches contain `Show`.

If the GD copy of `E01.mkv` has a different size, the status becomes `conflict`. `merge-immutable` refuses to overwrite it. Select the authoritative file manually before retrying.

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
