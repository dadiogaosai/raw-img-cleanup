package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dadiogaosai/rawtidy/cleanup"
)

func TestConfigFirstLaunchAndEditedParent(t *testing.T) {
	home := t.TempDir()
	config, err := loadConfig(home)
	if err != nil {
		t.Fatal(err)
	}
	if config.DefaultParent != home || config.LastJPEG != "" || config.LastRAW != "" {
		t.Fatalf("first launch config = %+v", config)
	}
	path := filepath.Join(home, ".config", "rawtidy", "config")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "default_parent: ") {
		t.Fatalf("not YAML config: %s", data)
	}
	parent := filepath.Join(home, "Photos")
	if err := os.WriteFile(path, []byte("default_parent: "+parent+"\nlast_jpeg: \"\"\nlast_raw: \"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err = loadConfig(home)
	if err != nil || config.DefaultParent != parent {
		t.Fatalf("edited config = %+v, %v", config, err)
	}
}

func TestConfigPathsAndSafeSave(t *testing.T) {
	home := t.TempDir()
	parent := filepath.Join(home, "Pictures")
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	canonicalParent, err := cleanup.CheckDirectory(parent)
	if err != nil {
		t.Fatal(err)
	}
	canonicalHome, err := cleanup.CheckDirectory(home)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfig(home)
	if err != nil {
		t.Fatal(err)
	}
	cfg.DefaultParent = "~/Pictures"
	cfg.LastJPEG = filepath.Join(home, "missing")
	if got, err := cfg.startDirectory(cfg.LastJPEG); err != nil || got != canonicalParent {
		t.Fatalf("missing saved directory fell back to %q, %v; want %q", got, err, canonicalParent)
	}
	if got, err := cfg.startDirectory("~/Pictures"); err != nil || got != canonicalParent {
		t.Fatalf("home-relative directory = %q, %v; want %q", got, err, canonicalParent)
	}
	cfg.DefaultParent = filepath.Join(home, "missing")
	if got, err := cfg.startDirectory(""); err != nil || got != canonicalHome {
		t.Fatalf("missing parent fell back to %q, %v; want home", got, err)
	}
	cfg.DefaultParent = "~/Pictures"
	cfg.LastJPEG = parent
	if err := cfg.save(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".config", "rawtidy", "config")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(before), "last_jpeg: "+parent) {
		t.Fatalf("saved config: %s", before)
	}
	cfg.LastJPEG = home
	backup := path + ".backup"
	if err := os.Rename(path, backup); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := cfg.save(); err == nil {
		t.Fatal("write over config directory succeeded")
	}
	after, err := os.ReadFile(backup)
	if err != nil || string(after) != string(before) {
		t.Fatalf("failed save changed config: %q, %v", after, err)
	}
}

func TestMalformedConfigIsNotReplaced(t *testing.T) {
	home := t.TempDir()
	if _, err := loadConfig(home); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".config", "rawtidy", "config")
	bad := []byte("default_parent: [oops\n")
	if err := os.WriteFile(path, bad, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(home); err == nil {
		t.Fatal("malformed YAML accepted")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(bad) {
		t.Fatalf("malformed config was replaced: %q, %v", got, err)
	}
}

func TestFirstLaunchDoesNotFollowExistingConfigSymlink(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "rawtidy")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, "unrelated")
	if err := os.Symlink(target, filepath.Join(dir, "config")); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(home); err == nil {
		t.Fatal("dangling config symlink accepted")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("first launch wrote through symlink: %v", err)
	}
}
