package agentui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSkillShowPrintsConfiguredSkill(t *testing.T) {
	cfg := Config{
		AgentName:    "fake-agent",
		SkillName:    "fake-agent",
		SkillContent: "---\nname: fake-agent\ndescription: fake\n---\n# Fake\n",
	}
	out, err := captureStdout(t, func() error {
		return Run(cfg, []string{"skill", "show"})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "name: fake-agent") {
		t.Errorf("expected output to contain 'name: fake-agent', got: %s", out)
	}
	if !strings.Contains(out, "# Fake") {
		t.Errorf("expected output to contain '# Fake', got: %s", out)
	}
}

func TestRunSkillInstallDryRunUsesConfiguredName(t *testing.T) {
	targetDir := t.TempDir()
	cfg := Config{
		AgentName:    "fake-agent",
		SkillName:    "fake-agent",
		SkillContent: "---\nname: fake-agent\n---\n# Fake\n",
	}
	out, err := captureStdout(t, func() error {
		return Run(cfg, []string{"skill", "install", "--dry-run", targetDir})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected output to contain '[dry-run]', got: %s", out)
	}
	if !strings.Contains(out, "SKILL.md") {
		t.Errorf("expected output to contain 'SKILL.md', got: %s", out)
	}
	_, statErr := os.Stat(filepath.Join(targetDir, "SKILL.md"))
	if statErr == nil {
		t.Error("expected no SKILL.md after dry run")
	}
}

func TestRunSkillInstallWritesSkillFile(t *testing.T) {
	targetDir := t.TempDir()
	skillContent := "---\nname: fake-agent\n---\n# Fake\n"
	cfg := Config{
		AgentName:    "fake-agent",
		SkillName:    "fake-agent",
		SkillContent: skillContent,
	}
	_, err := captureStdout(t, func() error {
		return Run(cfg, []string{"skill", "install", targetDir})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(targetDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	if string(data) != skillContent {
		t.Errorf("expected SKILL.md content %q, got %q", skillContent, string(data))
	}
}

func TestRunSkillUnknownSubcommand(t *testing.T) {
	cfg := Config{
		AgentName:    "fake-agent",
		SkillName:    "fake-agent",
		SkillContent: "---\nsome: content\n",
	}
	_, err := captureStdout(t, func() error {
		return Run(cfg, []string{"skill", "unknown"})
	})
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown") && !strings.Contains(err.Error(), "expected") {
		t.Errorf("expected error to mention 'unknown' or 'expected', got: %v", err)
	}
}

func TestRunSkillWithoutConfiguredContent(t *testing.T) {
	_, err := captureStdout(t, func() error {
		return Run(Config{AgentName: "fake-agent"}, []string{"skill", "show"})
	})
	if err == nil {
		t.Fatal("expected error when skill is not configured")
	}
	if !strings.Contains(err.Error(), "skill command is not configured") {
		t.Errorf("expected 'skill command is not configured', got: %v", err)
	}
}
