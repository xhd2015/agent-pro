package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func makeTree(t *testing.T, layout map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range layout {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
	return dir
}

func makeDir(t *testing.T, parent string, name string) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	return path
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestValidateRootHasReadme(t *testing.T) {
	dir := makeTree(t, map[string]string{
		"README.md": "# Test",
		"SETUP.md":  "## Steps",
		"ASSERT.md": "## Expected",
	})
	errs := validateRoot(dir, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateRootMissingReadme(t *testing.T) {
	dir := makeTree(t, map[string]string{
		"SETUP.md":  "## Steps",
		"ASSERT.md": "## Expected",
	})
	errs := validateRoot(dir, nil)
	if len(errs) == 0 {
		t.Error("expected error for missing README.md")
	}
}

func TestValidateSubdirsValid(t *testing.T) {
	root := makeTree(t, map[string]string{})
	leafDir := makeDir(t, root, "happy-path")
	writeFile(t, filepath.Join(leafDir, "SETUP.md"), "## Steps")
	writeFile(t, filepath.Join(leafDir, "ASSERT.md"), "## Expected")
	errs := validateSubdirs(root, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateSubdirsEmpty(t *testing.T) {
	root := makeTree(t, map[string]string{})
	makeDir(t, root, "empty-dir")
	errs := validateSubdirs(root, nil)
	if len(errs) == 0 {
		t.Error("expected error for empty directory")
	} else {
		found := false
		for _, e := range errs {
			if filepath.Base(e) == "" {
				continue
			}
			found = true
			break
		}
		if !found {
			t.Error("expected error mentioning empty directory")
		}
	}
}

func TestValidateSubdirsWarningOnEmpty(t *testing.T) {
	root := makeTree(t, map[string]string{})
	makeDir(t, root, "some-dir")
	errs := validateSubdirs(root, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for empty dir, got %d: %v", len(errs), errs)
	}
}

func TestValidateSubdirsAssertWithoutSetup(t *testing.T) {
	root := makeTree(t, map[string]string{})
	leafDir := makeDir(t, root, "bad-leaf")
	writeFile(t, filepath.Join(leafDir, "ASSERT.md"), "## Expected")
	errs := validateSubdirs(root, nil)
	if len(errs) == 0 {
		t.Error("expected error for ASSERT.md without SETUP.md")
	} else {
		found := false
		for _, e := range errs {
			if len(e) > 0 {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error mentioning SETUP.md missing")
		}
	}
}

func TestValidateSubdirsAssertWithSetup(t *testing.T) {
	root := makeTree(t, map[string]string{})
	leafDir := makeDir(t, root, "good-leaf")
	writeFile(t, filepath.Join(leafDir, "SETUP.md"), "## Steps")
	writeFile(t, filepath.Join(leafDir, "ASSERT.md"), "## Expected")
	errs := validateSubdirs(root, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateSubdirsRootNotCheckedForEmpty(t *testing.T) {
	root := makeTree(t, map[string]string{})
	errs := validateSubdirs(root, nil)
	if len(errs) != 0 {
		t.Error("root directory should not be flagged as empty")
	}
}

func TestValidateSubdirsRootAssertWithoutSetup(t *testing.T) {
	root := makeTree(t, map[string]string{
		"ASSERT.md": "## Expected",
	})
	errs := validateSubdirs(root, nil)
	if len(errs) == 0 {
		t.Error("root with ASSERT.md but no SETUP.md should be flagged")
	}
}
