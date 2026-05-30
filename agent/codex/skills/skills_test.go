package skills

import (
	"os"
	"path/filepath"
	"testing"

	agentskills "github.com/xhd2015/agent-pro/agent/skills"
)

func TestList_WithSkills(t *testing.T) {
	dir := t.TempDir()

	// Create codex-style skills
	createSkill(t, dir, "codex-skill", "codex-skill", "A codex skill")
	// Create agent-standard skills
	createSkill(t, dir, "agent-skill", "agent-skill", "An agent skill")

	// Use ListDir from shared package directly for unit testing
	skills, err := agentskills.ListDir(dir)
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d: %v", len(skills), skills)
	}

	names := map[string]bool{}
	for _, s := range skills {
		names[s.Name] = true
	}
	if !names["codex-skill"] || !names["agent-skill"] {
		t.Errorf("missing expected skills: %v", names)
	}
}

func TestList_SkipsDotDirs(t *testing.T) {
	dir := t.TempDir()

	systemDir := filepath.Join(dir, ".system", "some-skill")
	os.MkdirAll(systemDir, 0755)
	os.WriteFile(filepath.Join(systemDir, "SKILL.md"), []byte("---\nname: system-skill\ndescription: hidden\n---\n"), 0644)

	createSkill(t, dir, "visible", "visible", "Shown")

	skills, err := agentskills.ListDir(dir)
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "visible" {
		t.Errorf("expected visible, got %s", skills[0].Name)
	}
}

func TestList_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "plain")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`# Just markdown`), 0644)

	skills, err := agentskills.ListDir(dir)
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func createSkill(t *testing.T, root, dirName, name, description string) {
	t.Helper()
	skillDir := filepath.Join(root, dirName)
	os.MkdirAll(skillDir, 0755)
	content := "---\nname: " + name + "\ndescription: " + description + "\n---\n\n# " + name + "\n\nBody text."
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
}
