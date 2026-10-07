# RAW image cleanup

An interactive terminal app for moving RAW photos whose matching JPEG was deleted during review.

## Run

Requires Go 1.26 or newer. From this repository:

```sh
go run .
```

Or build a binary:

```sh
go build -o raw-img-cleanup .
./raw-img-cleanup
```

Choose the JPEG directory first, then the RAW directory. In the folder browser, use the arrow keys or `j`/`k` to highlight a folder, Enter or right arrow to open it, `h` or left arrow to go to its parent, and `s` to select the **current** folder. Press `e` to type or paste a path instead; Enter accepts it and Esc returns to browsing. Press `q` to quit. The app starts moving files as soon as both folders are valid.

## Matching and moves

The app scans both directory trees recursively. A RAW file is kept when a `.jpg` or `.jpeg` file with the exact same filename stem exists anywhere in the JPEG tree. Extension matching ignores letter case; stem matching does not. For example, `Trip/IMG_001.CR3` is kept if `Favorites/IMG_001.JPG` exists. Symbolic links are not followed.

Supported RAW extensions: `.arw`, `.cr2`, `.cr3`, `.crw`, `.dng`, `.nef`, `.nrw`, `.orf`, `.pef`, `.raf`, `.raw`, `.rw2`, `.rwl`, `.srw` (any letter case). Other files are left alone.

Unmatched RAW files move directly into `<RAW directory>/tmp-delete`. If that folder already contains the same filename, the app adds a numbered suffix, such as `IMG_001-1.CR3`; it never overwrites an existing file. `tmp-delete` is skipped on later runs. The result screen shows kept, moved, and failed counts plus paths for moved or failed files. Use up/down or `j`/`k` to scroll through long results.

Nothing is permanently deleted. To recover a photo, move it back from `tmp-delete` to the desired location. If the app reports a move failure, the source file remains in place.
