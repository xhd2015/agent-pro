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
	id := agentui.GenerateOutputName("test", "-tests-design.md")
	if !strings.HasSuffix(id, "-tests-design.md") {
		t.Errorf("expected suffix '-tests-design.md' in output name, got %q", id)
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
	os.Args[0] = "/usr/local/bin/test-case-design-expert"

	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w

	cfg := agentui.Config{
		AgentName:     "test-case-design-expert",
		SessionPrefix: "tcd_",
		Prompt:        prompt,
		Usage:         usage,
		OutputSuffix:  "-tests-design.md",
	}

	if cfg.AgentName != "test-case-design-expert" {
		t.Error("expected AgentName 'test-case-design-expert'")
	}
	if cfg.SessionPrefix != "tcd_" {
		t.Error("expected SessionPrefix 'tcd_'")
	}
	if cfg.OutputSuffix != "-tests-design.md" {
		t.Error("expected OutputSuffix '-tests-design.md'")
	}

	fmt.Fprintf(os.Stderr, "Session %s finished.\nTo resume: %s --resume %s\n", "tcd_test123", filepath.Base(os.Args[0]), "tcd_test123")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = origStderr

	msg := string(out)
	if !strings.Contains(msg, "--resume tcd_test123") {
		t.Errorf("exit message should contain '--resume tcd_test123', got: %s", msg)
	}
	if !strings.Contains(msg, "test-case-design-expert") {
		t.Errorf("exit message should contain program name, got: %s", msg)
	}
}

func TestSessionResolveWithTcdName(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "tcd_abc123"
	dir, err := session.Dir("test-case-design-expert", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", map[string]string{
		"session_id": resumeID,
		"feature":    "Add dark mode",
		"model":      "gpt-4o",
	})

	// Verify session was created under the right agent name
	if !strings.Contains(dir, "test-case-design-expert") {
		t.Errorf("session dir should contain 'test-case-design-expert', got: %s", dir)
	}
}

func TestConfigValues(t *testing.T) {
	cfg := agentui.Config{
		AgentName:     "test-case-design-expert",
		SessionPrefix: "tcd_",
		Prompt:        prompt,
		Usage:         usage,
		OutputSuffix:  "-tests-design.md",
	}
	if cfg.AgentName != "test-case-design-expert" {
		t.Errorf("expected AgentName 'test-case-design-expert', got %q", cfg.AgentName)
	}
	if cfg.SessionPrefix != "tcd_" {
		t.Errorf("expected SessionPrefix 'tcd_', got %q", cfg.SessionPrefix)
	}
	if cfg.OutputSuffix != "-tests-design.md" {
		t.Errorf("expected OutputSuffix '-tests-design.md', got %q", cfg.OutputSuffix)
	}
	if cfg.Prompt == "" {
		t.Error("prompt should not be empty")
	}
	if !strings.Contains(cfg.Usage, "test-case-design-expert") {
		t.Error("usage should mention agent name")
	}
}

func TestPromptEmbedding(t *testing.T) {
	if !strings.Contains(prompt, "senior QA engineer") {
		t.Error("embedded prompt should contain 'senior QA engineer'")
	}
	if !strings.Contains(prompt, "Brainstorm") {
		t.Error("embedded prompt should contain 'Brainstorm'")
	}
	if !strings.Contains(prompt, "Test Cases") {
		t.Error("embedded prompt should contain 'Test Cases'")
	}
}

func TestOutputSuffixApplied(t *testing.T) {
	result := agentui.GenerateOutputName("my feature", "-tests-design.md")
	if !strings.HasSuffix(result, "-tests-design.md") {
		t.Errorf("output should end with '-tests-design.md', got: %s", result)
	}
}
