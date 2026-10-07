package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Paths struct {
	JPEG string
	RAW  string
}

// Validate resolves and checks two separate, readable directory trees.
func Validate(jpeg, raw string) (Paths, error) {
	jpg, err := CheckDirectory(jpeg)
	if err != nil {
		return Paths{}, fmt.Errorf("JPEG directory: %w", err)
	}
	rawPath, err := CheckDirectory(raw)
	if err != nil {
		return Paths{}, fmt.Errorf("RAW directory: %w", err)
	}
	jpegContainsRAW, err := contains(jpg, rawPath)
	if err != nil {
		return Paths{}, fmt.Errorf("compare directories: %w", err)
	}
	rawContainsJPEG, err := contains(rawPath, jpg)
	if err != nil {
		return Paths{}, fmt.Errorf("compare directories: %w", err)
	}
	if jpegContainsRAW || rawContainsJPEG {
		return Paths{}, fmt.Errorf("JPEG and RAW directories overlap")
	}
	return Paths{JPEG: jpg, RAW: rawPath}, nil
}

// CheckDirectory returns the canonical path of an existing, readable directory.
func CheckDirectory(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path is empty")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", path)
	}
	if _, err := os.ReadDir(path); err != nil {
		return "", err
	}
	return path, nil
}

func contains(parent, child string) (bool, error) {
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return false, err
	}
	for current := child; ; current = filepath.Dir(current) {
		info, err := os.Stat(current)
		if err != nil {
			return false, err
		}
		if os.SameFile(parentInfo, info) {
			return true, nil
		}
		if filepath.Dir(current) == current {
			return false, nil
		}
	}
}
