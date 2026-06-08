package view

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildTreeClassifiesAndSortsDirectories(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "# Cases")
	writeTestFile(t, filepath.Join(root, "SETUP.md"), "## Global")
	writeTestFile(t, filepath.Join(root, "z-group", "ASSERT.md"), "## Expected")
	writeTestFile(t, filepath.Join(root, "a-group", "child", "ASSERT.md"), "## Expected")

	tree, err := buildTree(root)
	if err != nil {
		t.Fatalf("build tree: %v", err)
	}

	if tree.Type != "root" {
		t.Fatalf("root type = %q, want root", tree.Type)
	}
	if tree.Path != "" {
		t.Fatalf("root path = %q, want empty", tree.Path)
	}
	if len(tree.Children) != 2 {
		t.Fatalf("root children = %d, want 2", len(tree.Children))
	}
	if tree.Children[0].Name != "a-group" || tree.Children[1].Name != "z-group" {
		t.Fatalf("children not sorted by name: %q, %q", tree.Children[0].Name, tree.Children[1].Name)
	}
	if tree.Children[0].Type != "dir" {
		t.Fatalf("a-group type = %q, want dir", tree.Children[0].Type)
	}
	if tree.Children[1].Type != "leaf" {
		t.Fatalf("z-group type = %q, want leaf", tree.Children[1].Type)
	}
}

func TestSplitPathUsesSlashSeparatedSegments(t *testing.T) {
	got := splitPath("checkout/card/declined")
	want := []string{"checkout", "card", "declined"}
	if len(got) != len(want) {
		t.Fatalf("len(splitPath) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("splitPath[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
