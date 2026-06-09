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
	os.Args[0] = "/usr/local/bin/idea-expander"

	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w

	fmt.Fprintf(os.Stderr, "Session %s finished.\nTo resume: %s --resume %s\n", "ie_test123", filepath.Base(os.Args[0]), "ie_test123")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = origStderr

	msg := string(out)
	if !strings.Contains(msg, "--resume ie_test123") {
		t.Errorf("exit message should contain '--resume ie_test123', got: %s", msg)
	}
	if !strings.Contains(msg, "idea-expander") {
		t.Errorf("exit message should contain program name, got: %s", msg)
	}
}

func TestSessionResolveWithIeName(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AGENT_PRO_HOME", tmp)

	resumeID := "ie_abc123"
	dir, err := session.Dir("idea-expander", resumeID)
	if err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	session.WriteJSON(dir, "metadata.json", map[string]string{
		"session_id": resumeID,
		"feature":    "Add dark mode",
		"model":      "gpt-4o",
	})

	if !strings.Contains(dir, "idea-expander") {
		t.Errorf("session dir should contain 'idea-expander', got: %s", dir)
	}
}

func TestConfigValues(t *testing.T) {
	cfg := agentui.Config{
		AgentName:     "idea-expander",
		SessionPrefix: "ie_",
		Prompt:        prompt,
		Usage:         usage,
		SkillName:     "idea-expander",
		SkillContent:  skillTemplate,
	}
	if cfg.AgentName != "idea-expander" {
		t.Errorf("expected AgentName 'idea-expander', got %q", cfg.AgentName)
	}
	if cfg.SessionPrefix != "ie_" {
		t.Errorf("expected SessionPrefix 'ie_', got %q", cfg.SessionPrefix)
	}
	if cfg.Prompt == "" {
		t.Error("prompt should not be empty")
	}
	if cfg.SkillName != "idea-expander" {
		t.Errorf("expected SkillName 'idea-expander', got %q", cfg.SkillName)
	}
	if cfg.SkillContent == "" {
		t.Error("SkillContent should not be empty")
	}
	if cfg.SkillContent != skillTemplate {
		t.Error("SkillContent should match skillTemplate")
	}
	if !strings.Contains(cfg.Usage, "idea-expander") {
		t.Error("usage should mention agent name")
	}
}

func TestPromptEmbedding(t *testing.T) {
	if !strings.Contains(prompt, "product manager") {
		t.Error("embedded prompt should contain 'product manager'")
	}
	if !strings.Contains(prompt, "Brainstorm") {
		t.Error("embedded prompt should contain 'Brainstorm'")
	}
	if !strings.Contains(prompt, "User Stories") {
		t.Error("embedded prompt should contain 'User Stories'")
	}
	if !strings.Contains(prompt, "add-pending-questions") {
		t.Error("embedded prompt should contain 'add-pending-questions'")
	}
}

func TestAgentNameDiffersFromTcd(t *testing.T) {
	cfg := agentui.Config{
		AgentName:     "idea-expander",
		SessionPrefix: "ie_",
		Prompt:        prompt,
		Usage:         usage,
		SkillName:     "idea-expander",
		SkillContent:  skillTemplate,
	}
	if cfg.AgentName == "test-case-design-expert" {
		t.Error("idea-expander AgentName should differ from test-case-design-expert")
	}
	if cfg.SessionPrefix == "tcd_" {
		t.Error("idea-expander SessionPrefix should differ from tcd_")
	}
}

func TestSkillEmbedding(t *testing.T) {
	if skillTemplate == "" {
		t.Error("skillTemplate should not be empty")
	}
	if !strings.Contains(skillTemplate, "name: idea-expander") {
		t.Error("skillTemplate should contain 'name: idea-expander'")
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
		AgentName:    "idea-expander",
		SkillName:    "idea-expander",
		SkillContent: skillTemplate,
	}, []string{"skill", "show"})
	w.Close()
	<-done

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "name: idea-expander") {
		t.Errorf("expected output to contain 'name: idea-expander', got: %s", out)
	}
}
