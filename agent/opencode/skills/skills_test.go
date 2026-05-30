package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestList(t *testing.T) {
	projectDir := t.TempDir()
	localDir := filepath.Join(projectDir, ".opencode", "skills")
	os.MkdirAll(localDir, 0755)

	createSkill(t, localDir, "test-skill", "test-skill", "A test skill")

	// We can't reliably override the global home dir in tests,
	// so just test that the local skills are found.
	result, err := List(projectDir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Local) != 1 {
		t.Fatalf("expected 1 local skill, got %d", len(result.Local))
	}
	if result.Local[0].Name != "test-skill" {
		t.Errorf("name = %q", result.Local[0].Name)
	}
	if result.Local[0].Description != "A test skill" {
		t.Errorf("description = %q", result.Local[0].Description)
	}
}

func TestList_EmptyDir(t *testing.T) {
	projectDir := t.TempDir()
	// No .opencode/skills directory at all
	result, err := List(projectDir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Local) != 0 {
		t.Errorf("expected 0 local skills, got %d", len(result.Local))
	}
}

func TestList_SkipsDotDirs(t *testing.T) {
	projectDir := t.TempDir()
	localDir := filepath.Join(projectDir, ".opencode", "skills")
	os.MkdirAll(filepath.Join(localDir, ".hidden", "skill"), 0755)
	// Write a valid SKILL.md inside the hidden dir
	os.WriteFile(
		filepath.Join(localDir, ".hidden", "skill", "SKILL.md"),
		[]byte("---\nname: hidden-skill\ndescription: should not appear\n---\n"),
		0644,
	)

	createSkill(t, localDir, "shown", "shown", "Shown")

	result, err := List(projectDir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Local) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(result.Local))
	}
	if result.Local[0].Name != "shown" {
		t.Errorf("expected shown, got %s", result.Local[0].Name)
	}
}

func createSkill(t *testing.T, root, dirName, name, description string) {
	t.Helper()
	skillDir := filepath.Join(root, dirName)
	os.MkdirAll(skillDir, 0755)
	content := "---\nname: " + name + "\ndescription: " + description + "\n---\n\n# " + name + "\n"
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
}
