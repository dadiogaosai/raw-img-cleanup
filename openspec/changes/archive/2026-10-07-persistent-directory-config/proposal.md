## Why

The folder picker starts at the working directory and does not remember previous JPEG or RAW selections. Repeated cleanup runs require navigating to the same locations again.

## What Changes

- Add a YAML config at `~/.config/rawtidy/config`, created on first launch with the home directory as `default_parent`.
- Remember the last successfully selected JPEG and RAW directories separately.
- Open each selection step at its saved directory, falling back to `default_parent` and then home when a location is unavailable.
- Accept `~/` in configured paths, and report config parse or write failures rather than continuing with lost settings.

## Capabilities

### Modified Capabilities

- `interactive-raw-cleanup`: Directory selection gains persistent starting locations and a configurable fallback parent.

## Impact

The terminal picker and startup flow in `main.go`, its tests, and the README will change. YAML parsing may require a Go dependency; cleanup matching and movement stay as specified.
