# Interactive RAW Cleanup Specification

## Purpose

Help photographers separate unwanted RAW photos after reviewing the corresponding JPEG copies, while keeping the files recoverable and preventing accidental overwrites.

## Requirements

### Requirement: Select photo directories
The application SHALL let the user select a JPEG directory and then a RAW directory by browsing folders or entering paths. It SHALL reject missing, unreadable, or overlapping directories before changing files.

#### Scenario: Browse for directories
- **WHEN** the user browses to a JPEG folder and then a RAW folder
- **THEN** the application uses those folders for the cleanup run

#### Scenario: Enter directory paths
- **WHEN** the user enters valid JPEG and RAW directory paths
- **THEN** the application uses those folders for the cleanup run

#### Scenario: Invalid selection
- **WHEN** either directory is invalid or the selected trees overlap
- **THEN** the application displays an error and does not move files

### Requirement: Match RAW files to reviewed JPEG files
The application SHALL recursively scan regular files in both selected trees. JPEG candidates SHALL have `.jpg` or `.jpeg` extensions, ignoring extension case. RAW candidates SHALL have one of the documented supported RAW extensions, ignoring extension case. A RAW candidate SHALL be considered kept when any JPEG candidate anywhere in the JPEG tree has exactly the same filename stem, with stem case preserved. The application SHALL exclude the RAW root's `tmp-delete` subtree from scanning and SHALL not follow symbolic links.

#### Scenario: JPEG in a different subfolder
- **WHEN** a RAW file and a JPEG file have the same stem but appear in different relative subfolders
- **THEN** the RAW file is kept in place

#### Scenario: JPEG extension variants
- **WHEN** a RAW file's stem matches a file ending in `.jpg` or `.jpeg` in any letter case
- **THEN** the RAW file is kept in place

#### Scenario: Unsupported files and prior results
- **WHEN** the RAW tree contains unsupported file types or files already under its root `tmp-delete` directory
- **THEN** those files are not considered for movement

### Requirement: Move unmatched RAW files safely
The application SHALL start moving unmatched RAW candidates after valid directory selection without a separate confirmation step. It SHALL place them directly inside `<RAW directory>/tmp-delete`, creating that directory when needed. It SHALL never overwrite an existing destination file; when a name is occupied, it SHALL add a unique numbered suffix before the extension. It SHALL preserve source files that cannot be moved.

#### Scenario: Unmatched RAW file
- **WHEN** a supported RAW file has no matching JPEG stem
- **THEN** it is moved into the top level of `tmp-delete`

#### Scenario: Name collision
- **WHEN** `tmp-delete` already has the desired filename, including from another RAW subfolder
- **THEN** the moved file receives an unused numbered filename and the existing file remains unchanged

#### Scenario: Move failure
- **WHEN** an unmatched RAW file cannot be moved
- **THEN** the application reports the failure and leaves the source file in place

### Requirement: Report cleanup results
After a run, the application SHALL show the number of RAW candidates kept, moved, and failed, and SHALL identify any failed files. A run with no unmatched RAW files SHALL report that nothing was moved.

#### Scenario: Completed run
- **WHEN** the cleanup run finishes
- **THEN** the application displays kept, moved, and failed counts

#### Scenario: No unmatched files
- **WHEN** every RAW candidate has a matching JPEG stem
- **THEN** the application reports zero moved files
