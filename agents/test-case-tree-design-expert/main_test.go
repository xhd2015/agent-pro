package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/agentui"
	"github.com/xhd2015/agent-pro/agent/session"
)

func TestNewSessionIDViaAgentui(t *testing.T) {
	id := agentui.GenerateOutputName("test", "-test-cases")
	if !strings.HasSuffix(id, "-test-cases") {
		t.Errorf("expected suffix '-test-cases' in output name, got %q", id)
	}
}

func TestUsageContainsResumeFlag(t *testing.T) {
	if !strings.Contains(usage, "--resume") {
		t.Error("usage should mention --resume flag")
	}
	if !strings.Contains(usage, "Resume a previous session") {
		t.Error("usage should explain --resume")
	}
}

func TestExitMessageFormat(t *testing.T) {
	orig := os.Args[0]
	defer func() { os.Args[0] = orig }()
	os.Args[0] = "/usr/local/bin/test-case-tree-design-expert"

	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w

	cfg := agentui.Config{
		AgentName:     "test-case-tree-design-expert",
		SessionPrefix: "tctd_",
		Prompt:        prompt,
		Usage:         usage,
	}
	if cfg.AgentName != "test-case-tree-design-expert" {
		t.Error("expected AgentName 'test-case-tree-design-expert'")
	}
	if cfg.SessionPrefix != "tctd_" {
		t.Error("expected SessionPrefix 'tctd_'")
	}

	fmt.Fprintf(os.Stderr, "Session %s finished.\nTo resume: %s --resume %s\n", "tctd_test123", filepath.Base(os.Args[0]), "tctd_test123")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = origStderr

	msg := string(out)
	if !strings.Contains(msg, "--resume tctd_test123") {
		t.Errorf("exit message should contain '--resume tctd_test123', got: %s", msg)
	}
	if !strings.Contains(msg, "test-case-tree-design-expert") {
		t.Errorf("exit message should contain program name, got: %s", msg)
	}
}

func TestSessionResolveWithTctdName(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "tctd_abc123"
	dir, err := session.Dir("test-case-tree-design-expert", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", map[string]string{
		"session_id": resumeID,
		"feature":    "Add dark mode",
		"model":      "gpt-4o",
	})

	if !strings.Contains(dir, "test-case-tree-design-expert") {
		t.Errorf("session dir should contain 'test-case-tree-design-expert', got: %s", dir)
	}
}

func TestConfigValues(t *testing.T) {
	cfg := agentui.Config{
		AgentName:     "test-case-tree-design-expert",
		SessionPrefix: "tctd_",
		Prompt:        prompt,
		Usage:         usage,
	}
	if cfg.AgentName != "test-case-tree-design-expert" {
		t.Errorf("expected AgentName 'test-case-tree-design-expert', got %q", cfg.AgentName)
	}
	if cfg.SessionPrefix != "tctd_" {
		t.Errorf("expected SessionPrefix 'tctd_', got %q", cfg.SessionPrefix)
	}
	if cfg.Prompt == "" {
		t.Error("prompt should not be empty")
	}
	if !strings.Contains(cfg.Usage, "test-case-tree-design-expert") {
		t.Error("usage should mention agent name")
	}
}

func TestPromptEmbedding(t *testing.T) {
	if !strings.Contains(prompt, "senior QA engineer") {
		t.Error("embedded prompt should contain 'senior QA engineer'")
	}
	if !strings.Contains(prompt, "decision tree") {
		t.Error("embedded prompt should contain 'decision tree'")
	}
	if !strings.Contains(prompt, "SETUP.md") {
		t.Error("embedded prompt should contain 'SETUP.md'")
	}
	if !strings.Contains(prompt, "ASSERT.md") {
		t.Error("embedded prompt should contain 'ASSERT.md'")
	}
	if !strings.Contains(prompt, "mermaid") {
		t.Error("embedded prompt should contain 'mermaid'")
	}
	if !strings.Contains(prompt, "README.md") {
		t.Error("embedded prompt should contain 'README.md'")
	}
}

func TestOutputSuffixApplied(t *testing.T) {
	result := agentui.GenerateOutputName("my feature", "-test-cases")
	if !strings.HasSuffix(result, "-test-cases") {
		t.Errorf("output should end with '-test-cases', got: %s", result)
	}
}

func TestAgentNameDiffersFromTcd(t *testing.T) {
	cfg := agentui.Config{
		AgentName:     "test-case-tree-design-expert",
		SessionPrefix: "tctd_",
		Prompt:        prompt,
		Usage:         usage,
	}
	if cfg.AgentName == "test-case-design-expert" {
		t.Error("test-case-tree-design-expert AgentName should differ from test-case-design-expert")
	}
	if cfg.SessionPrefix == "tcd_" {
		t.Error("test-case-tree-design-expert SessionPrefix should differ from tcd_")
	}
	if cfg.AgentName == "idea-expander" {
		t.Error("test-case-tree-design-expert AgentName should differ from idea-expander")
	}
	if cfg.SessionPrefix == "ie_" {
		t.Error("test-case-tree-design-expert SessionPrefix should differ from ie_")
	}
}
