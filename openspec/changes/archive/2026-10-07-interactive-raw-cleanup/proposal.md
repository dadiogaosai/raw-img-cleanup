## Why

After reviewing and deleting unwanted JPEG photos, the matching RAW files remain in a separate directory. Finding and moving those RAW files by hand is slow and error prone, especially when photos are spread across subfolders.

## What Changes

- Add an interactive terminal application that lets the user browse for or enter the JPEG and RAW directories.
- Recursively identify RAW files whose filename stem has no matching `.jpg` or `.jpeg` file anywhere in the JPEG directory tree.
- Immediately move unmatched RAW files into `<RAW directory>/tmp-delete`, keeping all moved files at the top level and assigning unique names on collision without overwriting files.
- Report the results of the run and exclude `tmp-delete` from future scans.

## Capabilities

### New Capabilities

- `interactive-raw-cleanup`: Directory selection, recursive JPEG/RAW matching, safe movement of unmatched RAW files, and result reporting.

### Modified Capabilities

None.

## Impact

Adds a Go terminal application using Bubble Tea and filesystem operations. It reads the selected JPEG and RAW trees and moves unmatched RAW files within the RAW tree. No existing application code or API is changed.
