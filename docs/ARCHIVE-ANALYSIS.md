# Branch-aware archive analysis

Union filesystems such as mergerfs merge directory names and contents from multiple physical branches. A path visible at `/data/merged/Show` can contain some files from StorageBox and others from Google Drive. That merged path is useful to Sonarr and media players, but it cannot prove whether the source branch has been archived.

Atomic Sync always analyzes the physical job source and each configured physical destination. The comparison works for any supported directory-unit policy; mergerfs-backed media libraries are the specialized example in this guide. Do not configure the mergerfs union as both the source of truth and the archive destination.

## Decision model

The analyzer runs `rclone lsjson --recursive` once for the source and once for each destination. It groups entries using the job's directory-unit policy, then compares every source file with the same relative path at the unit's assigned destination.

```text
source has no files + destination has files       → archived
source has files + destination has no files       → pending
destination has files but misses source files     → partial
all source paths and sizes match                  → ready-to-verify
same relative path, different size or type        → conflict
files exist outside the assigned destination      → conflict
neither branch has files                           → empty
```

“Source has no files” includes an empty source directory shell. This matters with mergerfs: the shell can remain visible on StorageBox even though every real object lives on GD. The API separately reports `sourcePresent` and `destinationPresent`, so an operator can distinguish an absent unit from an empty shell without changing its macro status.

Destination-only extra files do not reduce source coverage. They may have been archived earlier or created by a media manager. Analysis never removes them. A unit can therefore be `partial` with 0% source coverage when the two physical branches contain entirely complementary files: the destination already holds part of the merged unit, but none of the files still present at the source has reached it.

An empty destination directory alone is not evidence of progress. If the source has files, the unit remains `pending`.

## What each state does—and does not—prove

| Status | Evidence in this snapshot | What it does not prove |
|---|---|---|
| `archived` | Destination files exist and no source files were listed | That a previous Atomic Sync run created them, or that the source mount is healthy |
| `ready-to-verify` | Every source relative path has the same size at the assigned destination | Equal file content or permission to delete the source |
| `partial` | Both sides contain content, but the destination does not cover all source paths | Whether a transfer is active, interrupted, or historically split |
| `pending` | Source has files and the destination has none | Whether the stable window has elapsed |
| `conflict` | At least one size/type or destination-placement invariant is violated | Which branch is authoritative |
| `empty` | Neither side contains files | Whether an empty shell is safe to remove |

`archived` is a conclusion about the current inventories, not historical proof of a completed move. A successful zero-file listing is valid after a real move, but the same result is possible when an existing mountpoint has lost its backing filesystem. Verify mount type and availability independently before trusting analysis.

The reference Compose model sets `create_host_path: false`, which prevents Docker from silently creating a missing bind source. It cannot distinguish a genuinely empty mounted share from an already-existing but unmounted empty directory.

Any source or destination listing error fails the complete analysis. Exhausted Drive quota is never converted into an empty inventory or a false `archived` result.

## Multiple destination branches

Existing SQLite assignments take precedence. An unassigned source unit uses the same deterministic weighted selection as execution without persisting a new assignment during analysis.

A destination-only unit is consolidated into one logical result on the branch that contains it. If files for that unit also exist on another configured destination, the result becomes `conflict`. Empty directory shells on secondary destinations remain presence metadata and do not create a false conflict.

Assignments are important with mergerfs: a file found somewhere in the merged archive is not necessarily on the branch selected for that unit. Atomic Sync reports content outside the selected branch rather than silently treating it as complete.

## Copy and move interpretation

Both product modes transfer directly to the assigned final path through rclone:

- after `copy`, a complete result normally appears as `ready-to-verify` because the source intentionally remains;
- after `move`, a complete result normally appears as `archived` because rclone has removed successfully moved source files;
- an interrupted or failed direct transfer can appear as `partial`, which is why same-named directories on both branches are never treated as success by themselves.

Move jobs may use either conflict policy and verification mode. `fail + size` is the fastest clean-destination canary; after reviewing a partial unit, `merge-immutable + checksum` provides stronger evidence for paths rclone actually transfers. Every move uses `--ignore-existing`, so destination-overlap paths are skipped rather than checksummed and still require independent content proof. Analysis itself remains metadata-first and does not perform transfer verification.

Atomic Sync does not expose rclone `sync`, so destination-only files are not pruned during either mode. Analysis is read-only and never authorizes a move or deletes content.

## Coverage and content verification

Dashboard coverage is the percentage of source file paths whose destination size matches. It is a low-I/O planning signal, not content-integrity proof. `archived` is displayed as 100% because no source files remain; this is a completion state, not evidence of how the destination files arrived.

Calculating checksums across a multi-terabyte CIFS mount merely to refresh a dashboard could disrupt imports and playback. Content verification is therefore explicit:

- `verify: size` compares paths and byte counts without reading file contents;
- `verify: checksum` enables rclone's `--checksum` transfer comparison and a hash supported by both backends when one is available;
- an independent `rclone check --download` is an optional deep audit that reads file contents from both sides and should be reserved for selected, quiesced units.

For a local or CIFS-mounted source and Google Drive, normal checksum transfer verification can calculate the local MD5 and compare it with Drive's stored MD5. It does not normally need to download the Drive object. Exact hash availability remains an rclone/backend property. Rclone owns this transfer verification, retries, and resumability. Every invocation is pinned to the discovery fingerprint's file paths through a temporary `--files-from-raw` manifest. After every non-dry-run copy or move, Atomic Sync separately requires each discovered file at the final destination with the same path and size; move then checks source residue. The manifest and metadata-completeness gate are not staging, a second content transfer, or `rclone check`.

Job validation rejects any source or destination endpoint whose normalized path contains a segment exactly named `.atomic-sync-staging`. A valid parent destination may still contain a legacy child with that name; destination inventory deliberately excludes the child. Version 0.2 never creates, transfers, or deletes that legacy recovery namespace; inventory and verify it separately before any explicit manual cleanup. Source discovery also fails closed if it encounters the namespace below an allowed source endpoint rather than silently treating it as payload.

## Analysis units versus executable units

The analyzer inventories shallow and scattered physical paths for operator review, but it does not attach a dedicated malformed-path diagnosis. Execution is stricter: a runnable unit must resolve to a directory at one fixed boundary:

- `folder`: one top-level directory for general directory trees;
- `depth`: exactly the configured positive number of directory components for custom general trees;
- `show`: one top-level show directory, a media preset equivalent to the `folder` boundary;
- `season`: exactly `Show/Season`, a two-level media preset.

The media presets are hierarchy labels only. They do not parse names, rename media, or determine whether a show or season is complete.

A file above that boundary is shallow. In particular, a loose file directly under `job.source` cannot be an executable unit; there is no per-file grouping mode. A discovery result containing both a parent directory and one of its descendants as separate units is ambiguous. Either condition fails the complete run before any rclone write starts. This prevents a root-level file and one of its related directories from being handled as independent transfer units.

## Examples

### Partially archived show

```text
StorageBox/Show/Season 01/E01.mkv
StorageBox/Show/Season 01/E02.mkv
GD/Show/Season 01/E01.mkv
```

mergerfs displays both episodes inside one `Show`. Atomic Sync reports `partial`, with one of two source files covered and one missing sample.

If GD's `E01.mkv` has a different size, the unit becomes `conflict`. The immutable merge policy will not overwrite it; select the authoritative copy before retrying.

### Completely complementary branches

```text
StorageBox/Movie/main.mkv
GD/Movie/poster.jpg
```

The merged directory contains both files, but GD contains none of the source payload. The result is `partial` with 0% source coverage—not `archived` and not `ready-to-verify`.

### Source retained after copy

```text
StorageBox/Movie/main.mkv
GD/Movie/main.mkv
```

When sizes match, analysis reports `ready-to-verify`. This is the expected physical state after copy mode; it does not silently reinterpret the duplicate as a move.

## Performance and API quota

- Analysis is read-only and metadata-first.
- Only one analysis runs at a time per Atomic Sync process.
- Destinations are listed sequentially to avoid quota bursts.
- Analysis and execution have separate concurrency limits.
- Persisted results let the UI display the latest snapshot without rescanning remote branches.

Avoid repeated manual refreshes while Drive reports quota exhaustion. Pause new scans, identify other consumers of the same OAuth project, wait for the quota window to recover, then resume one analysis at a time.

## Empty directories

Some object stores do not preserve empty directories, while CIFS does. Empty directory shells can inform branch analysis, but they are not entries in the transfer manifest and are not guaranteed to be copied or preserved. `empty` is informational. Do not create or delete content solely to make empty-directory counts match.
