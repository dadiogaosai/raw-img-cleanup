## Context

See [proposal.md](proposal.md) for motivation. `main.go` owns the two-step picker; `filepicker.New()` starts at `.` and the existing picker keeps its browsing position when JPEG selection advances to RAW. `cleanup.CheckDirectory` and `cleanup.Validate` already canonicalize and validate selections.

## Execution mode

**loop** — selected by the user on 2026-10-07 for test-first implementation and review gates across the config and picker tasks.

## Goals / Non-Goals

**Goals:** Keep persistence in one small config component, and use the existing directory validation for stored locations.

**Non-Goals:** Add in-app config editing or change cleanup matching and move behavior.

## Decisions

### Config format and location

Use the exact path `~/.config/rawtidy/config` and three YAML keys: `default_parent`, `last_jpeg`, and `last_raw`. Create the containing directory and config on first launch, with the resolved home path in `default_parent` and empty last-directory values. Use a YAML library rather than a partial hand-written parser, because the user selected YAML. JSON was considered for standard-library support but rejected by the user.

### Starting directories

Resolve absolute values and `~/` against the user's home directory. For each stage, check the saved location, then `default_parent`, then home with `cleanup.CheckDirectory`; the first valid directory becomes the picker's current directory. Reinitialize the picker when advancing to RAW so its listing reflects the RAW starting location. Keeping a single browsing position was considered but does not provide separate remembered locations.

### Saving and failure handling

Save the canonical path after each accepted selection; rejected selections and browsing do not write. Persist updates by writing a temporary file and renaming it over the config, so a failed write does not truncate the previous YAML. Report parse and write failures and stop the app. Keep the current config values, including a manually edited `default_parent`, when saving a last directory. Falling back silently from a malformed file was considered but rejected by the user.

## Risks / Trade-offs

- Re-encoding YAML can remove user comments or unsupported extra keys → Document the three supported keys and keep their values when saving.
- A saved path can disappear between startup validation and selection → Keep the existing selection-time validation and error display.

## Migration Plan

No existing config needs migration. First launch creates the file. Rolling back the binary leaves an unused config file and does not change photo files.
