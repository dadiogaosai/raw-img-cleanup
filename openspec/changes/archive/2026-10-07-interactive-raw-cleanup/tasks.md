## 1. Project foundation

- [x] 1.1 Create the Go module and terminal application entry point with Bubble Tea; verify `go build ./...` succeeds and the app starts.

## 2. Cleanup engine

- [x] 2.1 Add directory validation for existing, readable, nonoverlapping JPEG and RAW trees; verify temporary-directory tests cover valid paths, missing paths, and overlap before any move.
- [x] 2.2 Add recursive JPEG-stem and supported-RAW discovery, excluding symlinks and the RAW root's `tmp-delete`; verify tests cover extension case, nested folders, unsupported files, and exact stem matching.
- [x] 2.3 Move unmatched RAW files into a flat `tmp-delete` directory with numbered collision names and no overwrites; verify tests cover duplicate names, preexisting destinations, source preservation on failure, and rerunning cleanup.
- [x] 2.4 Return kept, moved, and failed counts plus per-file move results; verify tests cover an empty run and a partial-failure run.

## 3. Interactive workflow

- [x] 3.1 Add Bubble Tea folder browsing and editable path entry for JPEG selection followed by RAW selection; verify both input methods and invalid-path feedback in UI tests or an interactive terminal check.
- [x] 3.2 Start cleanup immediately after valid selection and display progress and final results, including moved source/destination paths and failures; verify a terminal run against temporary photo trees matches the cleanup engine's result.

## 4. Delivery

- [x] 4.1 Document how to build and run the app, folder selection, supported extensions, matching behavior, collision names, and recovery from `tmp-delete`; verify the documented commands work.
- [x] 4.2 Run `go test ./...`, `go vet ./...`, and `go build ./...`; resolve failures and verify the full application gate passes.
