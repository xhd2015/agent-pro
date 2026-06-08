package main

import (
	"os"
	"path/filepath"
	"testing"
)

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

	t.Run("wipes target before recreating", func(t *testing.T) {
		homeDir := t.TempDir()
		targetDir := t.TempDir()

		homeBase := filepath.Join(targetDir, "HOME")
		if err := os.MkdirAll(homeBase, 0755); err != nil {
			t.Fatal(err)
		}
		staleFile := filepath.Join(homeBase, "stale-file")
		if err := os.WriteFile(staleFile, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}

		createDir(t, homeDir, ".codex")
		err := CollectAgentFiles(homeDir, targetDir)
		if err != nil {
			t.Fatalf("CollectAgentFiles failed: %v", err)
		}

		assertPathNotExist(t, staleFile)
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
