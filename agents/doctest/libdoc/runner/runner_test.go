package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepoRoot(t *testing.T) {
	root := t.TempDir()
	writeRunnerTestFile(t, root, "go.mod", "module github.com/xhd2015/agent-pro\n")
	if err := os.MkdirAll(filepath.Join(root, "agents", "test-case-tree-runner"), 0755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "agents", "doctest", "tests")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}

	got, err := findRepoRoot()
	if err != nil {
		t.Fatalf("findRepoRoot: %v", err)
	}
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err = filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("root = %q, want %q", got, want)
	}
}

func TestFindRepoRootRejectsWrongModule(t *testing.T) {
	root := t.TempDir()
	writeRunnerTestFile(t, root, "go.mod", "module example.com/other\n")
	if err := os.MkdirAll(filepath.Join(root, "agents", "test-case-tree-runner"), 0755); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	_, err = findRepoRoot()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "could not locate agent-pro repository root") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFindRepoRootRequiresRunnerDirectory(t *testing.T) {
	root := t.TempDir()
	writeRunnerTestFile(t, root, "go.mod", "module github.com/xhd2015/agent-pro\n")

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	_, err = findRepoRoot()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "could not locate agent-pro repository root") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeRunnerTestFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
