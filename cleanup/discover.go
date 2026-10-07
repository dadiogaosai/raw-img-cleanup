package cleanup

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

var rawExtensions = map[string]bool{
	".arw": true, ".cr2": true, ".cr3": true, ".crw": true,
	".dng": true, ".nef": true, ".nrw": true, ".orf": true,
	".pef": true, ".raf": true, ".raw": true, ".rw2": true,
	".rwl": true, ".srw": true,
}

type Inventory struct {
	Kept      int
	Unmatched []string
}

// Discover finds RAW files without a JPEG with the same exact stem anywhere
// in the selected JPEG tree.
func Discover(paths Paths) (Inventory, error) {
	stems := make(map[string]bool)
	err := filepath.WalkDir(paths.JPEG, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") {
			stems[strings.TrimSuffix(name, ext)] = true
		}
		return nil
	})
	if err != nil {
		return Inventory{}, err
	}
	var result Inventory
	trash := filepath.Join(paths.RAW, "tmp-delete")
	err = filepath.WalkDir(paths.RAW, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == trash && entry.IsDir() {
			return filepath.SkipDir
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if !rawExtensions[strings.ToLower(ext)] {
			return nil
		}
		if stems[strings.TrimSuffix(name, ext)] {
			result.Kept++
		} else {
			result.Unmatched = append(result.Unmatched, path)
		}
		return nil
	})
	if err != nil {
		return Inventory{}, err
	}
	slices.Sort(result.Unmatched)
	return result, nil
}
