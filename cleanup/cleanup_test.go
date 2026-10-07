package cleanup

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestValidateDirectories(t *testing.T) {
	root := t.TempDir()
	jpg := filepath.Join(root, "jpg")
	raw := filepath.Join(root, "raw")
	for _, dir := range []string{jpg, raw} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Validate(jpg, raw); err != nil {
		t.Fatalf("separate directories: %v", err)
	}
	for _, tc := range []struct {
		name string
		jpg  string
		raw  string
	}{
		{"missing", jpg, filepath.Join(root, "missing")},
		{"same", jpg, jpg},
		{"nested", root, raw},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Validate(tc.jpg, tc.raw); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateRejectsCaseVariantOverlap(t *testing.T) {
	root := t.TempDir()
	jpeg := filepath.Join(root, "JPEG")
	if err := os.MkdirAll(filepath.Join(jpeg, "RAW"), 0o755); err != nil {
		t.Fatal(err)
	}
	variant := filepath.Join(root, "jpeg", "RAW")
	if _, err := os.Stat(variant); os.IsNotExist(err) {
		t.Skip("filesystem is case sensitive")
	} else if err != nil {
		t.Fatal(err)
	}
	if _, err := Validate(jpeg, variant); err == nil {
		t.Fatal("case variant allowed overlapping JPEG and RAW directories")
	}
}

func TestDiscoverMatchesAcrossFoldersAndSkipsPriorResults(t *testing.T) {
	root := t.TempDir()
	jpg := filepath.Join(root, "jpg")
	raw := filepath.Join(root, "raw")
	for _, dir := range []string{filepath.Join(jpg, "trip"), filepath.Join(raw, "other"), filepath.Join(raw, "lower"), filepath.Join(raw, "tmp-delete")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{
		filepath.Join(jpg, "trip", "Keep.JPEG"),
		filepath.Join(raw, "other", "Keep.CR3"),
		filepath.Join(raw, "lower", "keep.CR3"),
		filepath.Join(raw, "other", "Drop.NeF"),
		filepath.Join(raw, "other", "note.txt"),
		filepath.Join(raw, "tmp-delete", "Old.CR3"),
	} {
		if err := os.WriteFile(path, []byte("photo"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(raw, "other", "Drop.NeF"), filepath.Join(raw, "linked.NeF")); err != nil {
		t.Fatal(err)
	}
	paths, err := Validate(jpg, raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Discover(paths)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kept != 1 {
		t.Fatalf("kept = %d, want 1", got.Kept)
	}
	want := []string{filepath.Join(paths.RAW, "lower", "keep.CR3"), filepath.Join(paths.RAW, "other", "Drop.NeF")}
	if !slices.Equal(got.Unmatched, want) {
		t.Fatalf("unmatched = %v, want %v", got.Unmatched, want)
	}
}

func TestMoveUnmatchedFlattensWithoutOverwriting(t *testing.T) {
	root := t.TempDir()
	raw := filepath.Join(root, "raw")
	for _, dir := range []string{filepath.Join(raw, "a"), filepath.Join(raw, "b"), filepath.Join(raw, "tmp-delete")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	first := filepath.Join(raw, "a", "Same.CR3")
	second := filepath.Join(raw, "b", "Same.CR3")
	prior := filepath.Join(raw, "tmp-delete", "Same.CR3")
	for path, content := range map[string]string{first: "first", second: "second", prior: "prior"} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	result := Move(Paths{RAW: raw}, Inventory{Unmatched: []string{first, second}})
	if result.Moved != 2 || result.Failed != 0 {
		t.Fatalf("result = %+v", result)
	}
	for path, want := range map[string]string{
		prior: "prior",
		filepath.Join(raw, "tmp-delete", "Same-1.CR3"): "first",
		filepath.Join(raw, "tmp-delete", "Same-2.CR3"): "second",
	} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("%s = %q, %v; want %q", path, got, err, want)
		}
	}
	for _, path := range []string{first, second} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("source still exists: %s", path)
		}
	}
}

func TestMoveFailurePreservesSourceAndRerunSkipsTrash(t *testing.T) {
	root := t.TempDir()
	jpg := filepath.Join(root, "jpg")
	raw := filepath.Join(root, "raw")
	for _, dir := range []string{jpg, raw} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(raw, "Only.CR3")
	if err := os.WriteFile(source, []byte("raw"), 0o644); err != nil {
		t.Fatal(err)
	}
	trash := filepath.Join(raw, "tmp-delete")
	if err := os.WriteFile(trash, []byte("blocker"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths, err := Validate(jpg, raw)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := Discover(paths)
	if err != nil {
		t.Fatal(err)
	}
	failed := Move(paths, inventory)
	if failed.Failed != 1 || failed.Moved != 0 {
		t.Fatalf("blocked move = %+v", failed)
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatalf("source lost after failure: %v", err)
	}
	if err := os.Remove(trash); err != nil {
		t.Fatal(err)
	}
	if moved := Move(paths, inventory); moved.Moved != 1 || moved.Failed != 0 {
		t.Fatalf("retry = %+v", moved)
	}
	again, err := Discover(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Unmatched) != 0 {
		t.Fatalf("rerun found prior result: %v", again.Unmatched)
	}
}

func TestRunEmptyAndPartialMoveResults(t *testing.T) {
	root := t.TempDir()
	jpg := filepath.Join(root, "jpg")
	raw := filepath.Join(root, "raw")
	for _, dir := range []string{jpg, raw} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	empty, err := Run(jpg, raw)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Kept != 0 || empty.Moved != 0 || empty.Failed != 0 {
		t.Fatalf("empty result = %+v", empty)
	}
	good := filepath.Join(raw, "Good.CR3")
	if err := os.WriteFile(good, []byte("raw"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths, err := Validate(jpg, raw)
	if err != nil {
		t.Fatal(err)
	}
	partial := Move(paths, Inventory{Kept: 2, Unmatched: []string{filepath.Join(raw, "missing.CR3"), good}})
	if partial.Kept != 2 || partial.Moved != 1 || partial.Failed != 1 || len(partial.Files) != 1 || len(partial.Failures) != 1 {
		t.Fatalf("partial result = %+v", partial)
	}
}
