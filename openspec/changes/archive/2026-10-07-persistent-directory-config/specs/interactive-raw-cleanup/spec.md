## MODIFIED Requirements

### Requirement: Select photo directories
The application SHALL let the user select a JPEG directory and then a RAW directory by browsing folders or entering paths. It SHALL reject missing, unreadable, or overlapping directories before changing files. For each selection step, the browser SHALL start at the last successfully selected directory of that type when available, otherwise at the configured default parent when available, otherwise at the user's home directory.

#### Scenario: Browse for directories
- **WHEN** the user browses to a JPEG folder and then a RAW folder
- **THEN** the application uses those folders for the cleanup run

#### Scenario: Enter directory paths
- **WHEN** the user enters valid JPEG and RAW directory paths
- **THEN** the application uses those folders for the cleanup run

#### Scenario: Invalid selection
- **WHEN** either directory is invalid or the selected trees overlap
- **THEN** the application displays an error and does not move files

#### Scenario: Resume separate selection locations
- **WHEN** the user opens the JPEG or RAW selection step after previously selecting a directory of that type
- **THEN** the browser starts in that type's saved directory, if it is available

#### Scenario: Fall back from unavailable locations
- **WHEN** a saved directory is unavailable or empty
- **THEN** the browser starts in the configured default parent if available, or home otherwise

## ADDED Requirements

### Requirement: Persist directory preferences
The application SHALL maintain a YAML file at `~/.config/rawtidy/config` containing `default_parent`, `last_jpeg`, and `last_raw`. On first launch, it SHALL create the file with `default_parent` set to the user's home directory. It SHALL accept absolute paths and `~/` home-relative paths in the file. It SHALL update `last_jpeg` or `last_raw` only after a directory of that type is successfully selected. It SHALL report a config parse or write failure and stop rather than silently discard saved preferences.

#### Scenario: First launch
- **WHEN** the user starts the application without a config file
- **THEN** the application creates a YAML config with home as `default_parent` and starts JPEG selection there

#### Scenario: Save accepted selections
- **WHEN** the user successfully selects a JPEG or RAW directory
- **THEN** the corresponding last-directory value is saved for a later launch

#### Scenario: Invalid selection does not replace saved path
- **WHEN** the user browses without selecting or attempts to select an invalid directory
- **THEN** the previously saved directory values remain unchanged

#### Scenario: Home-relative configured path
- **WHEN** a configured directory starts with `~/`
- **THEN** the application resolves it under the user's home directory

#### Scenario: Config failure
- **WHEN** the config cannot be parsed or written
- **THEN** the application reports the error and stops without replacing the existing config with defaults
