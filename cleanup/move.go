package cleanup

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type MovedFile struct {
	Source      string
	Destination string
}

type Failure struct {
	Source string
	Err    error
}

type Result struct {
	Kept     int
	Moved    int
	Failed   int
	Files    []MovedFile
	Failures []Failure
}

// Move places unmatched RAW files in a flat tmp-delete directory without
// replacing any existing destination. Linking first makes the name claim
// atomic; removing the original then completes the move.
func Move(paths Paths, inventory Inventory) Result {
	result := Result{Kept: inventory.Kept}
	if len(inventory.Unmatched) == 0 {
		return result
	}
	trash := filepath.Join(paths.RAW, "tmp-delete")
	if err := os.Mkdir(trash, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
		for _, source := range inventory.Unmatched {
			result.fail(source, err)
		}
		return result
	}
	info, err := os.Lstat(trash)
	if err != nil || !info.IsDir() {
		if err == nil {
			err = fmt.Errorf("tmp-delete is not a directory")
		}
		for _, source := range inventory.Unmatched {
			result.fail(source, err)
		}
		return result
	}
	for _, source := range inventory.Unmatched {
		destination, err := moveOne(source, trash)
		if err != nil {
			result.fail(source, err)
			continue
		}
		result.Moved++
		result.Files = append(result.Files, MovedFile{Source: source, Destination: destination})
	}
	return result
}

// Run validates the selected trees, discovers candidates, and moves files.
func Run(jpegDir, rawDir string) (Result, error) {
	paths, err := Validate(jpegDir, rawDir)
	if err != nil {
		return Result{}, err
	}
	inventory, err := Discover(paths)
	if err != nil {
		return Result{}, fmt.Errorf("scan photo directories: %w", err)
	}
	return Move(paths, inventory), nil
}

func moveOne(source, trash string) (string, error) {
	info, err := os.Lstat(source)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("source is no longer a regular file")
	}
	name := filepath.Base(source)
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for n := 0; ; n++ {
		candidate := name
		if n > 0 {
			candidate = fmt.Sprintf("%s-%d%s", stem, n, ext)
		}
		destination := filepath.Join(trash, candidate)
		err := os.Link(source, destination)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			// Filesystems such as exFAT may not support hard links.
			err = copyExclusive(source, destination, info)
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			if err != nil {
				return "", err
			}
		}
		if err := os.Remove(source); err != nil {
			if rollbackErr := os.Remove(destination); rollbackErr != nil {
				return "", fmt.Errorf("remove source: %w; rollback destination: %v", err, rollbackErr)
			}
			return "", fmt.Errorf("remove source: %w", err)
		}
		return destination, nil
	}
}

func copyExclusive(source, destination string, info fs.FileInfo) (err error) {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.Remove(destination)
		}
	}()
	if _, err = io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	if err = output.Sync(); err != nil {
		output.Close()
		return err
	}
	if err = output.Close(); err != nil {
		return err
	}
	if err = os.Chmod(destination, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Chtimes(destination, info.ModTime(), info.ModTime())
}

func (r *Result) fail(source string, err error) {
	r.Failed++
	r.Failures = append(r.Failures, Failure{Source: source, Err: err})
}
