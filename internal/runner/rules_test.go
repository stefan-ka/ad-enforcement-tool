package runner

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestCollectRuleFiles_SingleFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "my.rule")
	if err := os.WriteFile(f, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := CollectRuleFiles(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != f {
		t.Errorf("got %v, want [%s]", got, f)
	}
}

func TestCollectRuleFiles_NonExistentPath(t *testing.T) {
	_, err := CollectRuleFiles("/definitely/does/not/exist.rule")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestCollectRuleFiles_DirectoryCollectsRuleFiles(t *testing.T) {
	dir := t.TempDir()
	names := []string{"a.rule", "b.rule", "c.rule"}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// add a non-.rule file that must be ignored
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := CollectRuleFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d files, want 3: %v", len(got), got)
	}
	sort.Strings(got)
	for i, n := range names {
		if filepath.Base(got[i]) != n {
			t.Errorf("got[%d] = %q, want %q", i, got[i], n)
		}
	}
}

func TestCollectRuleFiles_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	got, err := CollectRuleFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestCollectRuleFiles_RecursiveSubdirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	topFile := filepath.Join(dir, "top.rule")
	subFile := filepath.Join(sub, "nested.rule")
	for _, f := range []string{topFile, subFile} {
		if err := os.WriteFile(f, []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := CollectRuleFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d files, want 2: %v", len(got), got)
	}
}
