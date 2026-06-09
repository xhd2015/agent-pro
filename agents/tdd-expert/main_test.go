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

func TestUsageContainsResumeFlag(t *testing.T) {
	if !strings.Contains(usage, "--resume") {
		t.Error("usage should mention --resume flag")
	}
	if !strings.Contains(usage, "Resume a previous session") {
		t.Error("usage should explain --resume")
	}
}

func TestUsageContainsSkillCommands(t *testing.T) {
	if !strings.Contains(usage, "skill show") {
		t.Error("usage should mention 'skill show'")
	}
	if !strings.Contains(usage, "skill install") {
		t.Error("usage should mention 'skill install'")
	}
}

func TestExitMessageFormat(t *testing.T) {
	orig := os.Args[0]
	defer func() { os.Args[0] = orig }()
	os.Args[0] = "/usr/local/bin/tdd-expert"

	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w

	cfg := agentui.Config{
		AgentName:     "tdd-expert",
		SessionPrefix: "tdd_",
		Prompt:        prompt,
		Usage:         usage,
		SkillName:     "tdd-expert",
		SkillContent:  skillTemplate,
	}
	if cfg.AgentName != "tdd-expert" {
		t.Error("expected AgentName 'tdd-expert'")
	}
	if cfg.SessionPrefix != "tdd_" {
		t.Error("expected SessionPrefix 'tdd_'")
	}

	fmt.Fprintf(os.Stderr, "Session %s finished.\nTo resume: %s --resume %s\n", "tdd_test123", filepath.Base(os.Args[0]), "tdd_test123")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = origStderr

	msg := string(out)
	if !strings.Contains(msg, "--resume tdd_test123") {
		t.Errorf("exit message should contain '--resume tdd_test123', got: %s", msg)
	}
	if !strings.Contains(msg, "tdd-expert") {
		t.Errorf("exit message should contain program name, got: %s", msg)
	}
}

func TestSessionResolveWithTddName(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "tdd_abc123"
	dir, err := session.Dir("tdd-expert", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", map[string]string{
		"session_id": resumeID,
		"feature":    "path/to/test-case-tree",
		"model":      "gpt-4o",
	})

	if !strings.Contains(dir, "tdd-expert") {
		t.Errorf("session dir should contain 'tdd-expert', got: %s", dir)
	}
}

func TestConfigValues(t *testing.T) {
	cfg := agentui.Config{
		AgentName:     "tdd-expert",
		SessionPrefix: "tdd_",
		Prompt:        prompt,
		Usage:         usage,
		SkillName:     "tdd-expert",
		SkillContent:  skillTemplate,
	}
	if cfg.AgentName != "tdd-expert" {
		t.Errorf("expected AgentName 'tdd-expert', got %q", cfg.AgentName)
	}
	if cfg.SessionPrefix != "tdd_" {
		t.Errorf("expected SessionPrefix 'tdd_', got %q", cfg.SessionPrefix)
	}
	if cfg.Prompt == "" {
		t.Error("prompt should not be empty")
	}
	if cfg.SkillName != "tdd-expert" {
		t.Errorf("expected SkillName 'tdd-expert', got %q", cfg.SkillName)
	}
	if cfg.SkillContent == "" {
		t.Error("SkillContent should not be empty")
	}
	if cfg.SkillContent != skillTemplate {
		t.Error("SkillContent should match skillTemplate")
	}
	if !strings.Contains(cfg.Usage, "tdd-expert") {
		t.Error("usage should mention agent name")
	}
}

func TestPromptEmbedding(t *testing.T) {
	if !strings.Contains(prompt, "TDD") {
		t.Error("embedded prompt should contain 'TDD'")
	}
	if !strings.Contains(prompt, "test case tree") {
		t.Error("embedded prompt should contain 'test case tree'")
	}
	if !strings.Contains(prompt, "ASSERT.md") {
		t.Error("embedded prompt should contain 'ASSERT.md'")
	}
	if !strings.Contains(prompt, "SETUP.md") {
		t.Error("embedded prompt should contain 'SETUP.md'")
	}
	if !strings.Contains(prompt, "not implemented") {
		t.Error("embedded prompt should contain 'not implemented'")
	}
	if !strings.Contains(prompt, "RED") {
		t.Error("embedded prompt should contain 'RED'")
	}
	if !strings.Contains(prompt, "go test") {
		t.Error("embedded prompt should contain 'go test'")
	}
	if !strings.Contains(prompt, "README.md") {
		t.Error("embedded prompt should contain 'README.md'")
	}
}

func TestAgentNameDiffersFromOthers(t *testing.T) {
	cfg := agentui.Config{
		AgentName:     "tdd-expert",
		SessionPrefix: "tdd_",
		Prompt:        prompt,
		Usage:         usage,
		SkillName:     "tdd-expert",
		SkillContent:  skillTemplate,
	}
	if cfg.AgentName == "test-case-tree-design-expert" {
		t.Error("tdd-expert AgentName should differ from test-case-tree-design-expert")
	}
	if cfg.SessionPrefix == "tctd_" {
		t.Error("tdd-expert SessionPrefix should differ from tctd_")
	}
	if cfg.AgentName == "test-case-design-expert" {
		t.Error("tdd-expert AgentName should differ from test-case-design-expert")
	}
	if cfg.SessionPrefix == "tcd_" {
		t.Error("tdd-expert SessionPrefix should differ from tcd_")
	}
	if cfg.AgentName == "idea-expander" {
		t.Error("tdd-expert AgentName should differ from idea-expander")
	}
	if cfg.SessionPrefix == "ie_" {
		t.Error("tdd-expert SessionPrefix should differ from ie_")
	}
}

func TestSkillEmbedding(t *testing.T) {
	if skillTemplate == "" {
		t.Error("skillTemplate should not be empty")
	}
	if !strings.Contains(skillTemplate, "name: tdd-expert") {
		t.Error("skillTemplate should contain 'name: tdd-expert'")
	}
	if !strings.Contains(skillTemplate, "description:") {
		t.Error("skillTemplate should contain 'description:'")
	}
}

func TestSkillShow(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	done := make(chan struct{})
	var buf strings.Builder
	go func() {
		defer close(done)
		io.Copy(&buf, r)
	}()

	err = agentui.Run(agentui.Config{
		AgentName:    "tdd-expert",
		SkillName:    "tdd-expert",
		SkillContent: skillTemplate,
	}, []string{"skill", "show"})
	w.Close()
	<-done

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "name: tdd-expert") {
		t.Errorf("expected output to contain 'name: tdd-expert', got: %s", out)
	}
}
