## 1. Config lifecycle

- [x] 1.1 Add YAML config loading and first-launch creation at `~/.config/rawtidy/config`; verify a test creates the file with home as `default_parent` and reads an edited value.
- [x] 1.2 Add path expansion, fallback validation, and safe saving of last-directory values; verify tests cover `~/`, unavailable directories, malformed YAML, and write failures without losing the previous config.

## 2. Directory selection

- [x] 2.1 Start JPEG and RAW pickers at their separate saved locations with parent/home fallbacks; verify a UI test sees the expected directory at each stage.
- [x] 2.2 Save only successfully accepted selections and stop on config write errors before cleanup; verify a UI test covers invalid selection, persistence, and the failure path.

## 3. Documentation and integration

- [x] 3.1 Document the YAML keys, first-launch behavior, and directory precedence in `README.md`; verify the example matches the accepted config format.
- [x] 3.2 Run `go test ./...` and verify the complete suite passes.
