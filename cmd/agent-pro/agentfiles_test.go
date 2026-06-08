package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateTargetDir(t *testing.T) {
	t.Run("accepts non-existent dir", func(t *testing.T) {
		err := ValidateTargetDir(filepath.Join(t.TempDir(), "does-not-exist"))
		if err != nil {
			t.Errorf("non-existent dir should be valid: %v", err)
		}
	})

	t.Run("accepts empty dir", func(t *testing.T) {
		err := ValidateTargetDir(t.TempDir())
		if err != nil {
			t.Errorf("empty dir should be valid: %v", err)
		}
	})

	t.Run("rejects non-empty dir", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "stale"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		err := ValidateTargetDir(dir)
		if err == nil {
			t.Fatal("non-empty dir should be rejected")
		}
	})

	t.Run("rejects file path", func(t *testing.T) {
		f, err := os.CreateTemp("", "agent-pro-test")
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		defer os.Remove(f.Name())

		err = ValidateTargetDir(f.Name())
		if err == nil {
			t.Fatal("file path should be rejected")
		}
	})
}

func TestCollectAgentFiles(t *testing.T) {
	t.Run("creates symlinks for existing agent paths", func(t *testing.T) {
		homeDir := t.TempDir()
		targetDir := t.TempDir()

		createDir(t, homeDir, ".codex")
		createDir(t, homeDir, ".claude")
		createDir(t, homeDir, ".config", "opencode")
		createDir(t, homeDir, ".gemini")

		err := CollectAgentFiles(homeDir, targetDir)
		if err != nil {
			t.Fatalf("CollectAgentFiles failed: %v", err)
		}

		homeBase := filepath.Join(targetDir, "HOME")

		checkSymlink(t, homeDir, homeBase, ".codex")
		checkSymlink(t, homeDir, homeBase, ".claude")
		checkSymlink(t, homeDir, homeBase, ".config", "opencode")
		checkSymlink(t, homeDir, homeBase, ".gemini")

		assertPathNotExist(t, filepath.Join(homeBase, ".agents"))
		assertPathNotExist(t, filepath.Join(homeBase, ".cursor"))
	})

	t.Run("skips non-existing paths", func(t *testing.T) {
		homeDir := t.TempDir()
		targetDir := t.TempDir()

		err := CollectAgentFiles(homeDir, targetDir)
		if err != nil {
			t.Fatalf("CollectAgentFiles failed: %v", err)
		}

		homeBase := filepath.Join(targetDir, "HOME")
		entries, err := os.ReadDir(homeBase)
		if err != nil {
			t.Fatalf("read HOME dir: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected no entries, got %d", len(entries))
		}
	})

	t.Run("link targets resolve correctly", func(t *testing.T) {
		homeDir := t.TempDir()
		targetDir := t.TempDir()

		marker := filepath.Join(homeDir, ".codex", "marker.txt")
		createDir(t, homeDir, ".codex")
		if err := os.WriteFile(marker, []byte("agent data"), 0644); err != nil {
			t.Fatal(err)
		}

		err := CollectAgentFiles(homeDir, targetDir)
		if err != nil {
			t.Fatalf("CollectAgentFiles failed: %v", err)
		}

		resolved, err := os.ReadFile(filepath.Join(targetDir, "HOME", ".codex", "marker.txt"))
		if err != nil {
			t.Fatalf("read through symlink: %v", err)
		}
		if string(resolved) != "agent data" {
			t.Errorf("expected 'agent data', got '%s'", string(resolved))
		}
	})

	t.Run("handles non-existent home dir gracefully", func(t *testing.T) {
		err := CollectAgentFiles("/nonexistent/home/dir", t.TempDir())
		if err != nil {
			t.Fatalf("should not error on non-existent home dir: %v", err)
		}
	})
}

func createDir(t *testing.T, parts ...string) string {
	t.Helper()
	dir := filepath.Join(parts...)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func checkSymlink(t *testing.T, homeDir string, base string, relParts ...string) {
	t.Helper()

	linkPath := filepath.Join(append([]string{base}, relParts...)...)

	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Errorf("symlink missing: %s", linkPath)
		return
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("not a symlink: %s", linkPath)
		return
	}

	resolved, err := os.Readlink(linkPath)
	if err != nil {
		t.Errorf("readlink failed for %s: %v", linkPath, err)
		return
	}

	expected := filepath.Join(append([]string{homeDir}, relParts...)...)
	if resolved != expected {
		t.Errorf("symlink %s points to %s, want %s", linkPath, resolved, expected)
	}
}

func assertPathNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("path %s exists but should not", path)
	}
}
