## Context

The repository currently has OpenSpec configuration and no application code. See [proposal.md](proposal.md) for the user workflow. The new application needs a terminal UI and filesystem traversal, with all moves confined to the selected RAW tree.

## Execution mode

**implement** — selected on 2026-10-07. This change combines testable filesystem behavior with terminal UI wiring. The requested `loop` mode cannot run because OCR has no configured LLM endpoint; the user chose to continue in `implement` mode.

## Goals / Non-Goals

**Goals:**
- Make directory choice and cleanup usable entirely from a terminal.
- Keep matching and movement logic separate from the UI so it can be tested with temporary directories.
- Preserve recoverability and avoid overwriting any file.

**Non-Goals:**
- Delete files permanently or modify JPEG files.
- Compare image contents or metadata; filename stems are the sole matching key.
- Add a persistent database or background watcher.

## Decisions

### Terminal interaction

Use Bubble Tea for a stateful flow: choose JPEG directory, choose RAW directory, run cleanup, display results. Provide both a folder browser and editable path input at each selection step. Validate both canonical directory paths before starting the run, rejecting identical or ancestor/descendant selections. Run scanning and movement through a Bubble Tea command so the UI remains responsive. A path-only form was considered, but the confirmed workflow calls for browsing too.

### File discovery and matching

Use `filepath.WalkDir` to collect JPEG stems into a set, then walk the RAW tree. Consider `.jpg` and `.jpeg` as JPEG extensions. Start with an explicit, documented common RAW extension set: `.arw`, `.cr2`, `.cr3`, `.crw`, `.dng`, `.nef`, `.nrw`, `.orf`, `.pef`, `.raf`, `.raw`, `.rw2`, `.rwl`, and `.srw`. Extension comparison is case-insensitive; stem comparison is exact. Skip symbolic links, non-regular files, and the RAW root's `tmp-delete` subtree. Using relative paths as matching keys was considered, but the user chose matching stems anywhere in the JPEG tree.

### Move behavior

Create `tmp-delete` as needed and move every unmatched RAW candidate to its top level. Reserve an available filename using a numbered suffix before the extension (`IMG_001-1.CR3`, etc.) when the original name is occupied. Use a no-overwrite move strategy so a destination created concurrently cannot be replaced. Continue after an individual move error, retain that source, and collect failures for the results screen. Preserving source subfolders in `tmp-delete` was considered, but the user chose a flat destination. No preview step is added because immediate movement was confirmed.

## Risks / Trade-offs

- A JPEG with the same stem in an unrelated subfolder will keep a RAW file. This follows the selected matching rule; show the rule in the UI and documentation.
- The explicit extension list may omit a camera format. Document the list so users can check support and extend it later.
- A permission or filesystem error can leave a partial run. Report each failed source and its reason; successful moves stay in `tmp-delete` for manual recovery.
- Flattening subfolders loses their original path in the destination. Collision suffixes preserve every file, but the result view should include source and destination paths for moved files.

## Migration Plan

No migration is needed. Introduce the Go module and application, document how to run it, and keep the initial version local to the selected directories.
