package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	if err := handleSkillCommand([]string{"show"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Close()
	<-done

	out := buf.String()
	if !strings.Contains(out, "name: llm-mock") {
		t.Errorf("expected output to contain 'name: llm-mock', got: %s", out)
	}
}

func TestSkillInstallDryRun(t *testing.T) {
	targetDir := t.TempDir()
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

	if err := handleSkillCommand([]string{"install", "--dry-run", targetDir}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Close()
	<-done

	out := buf.String()
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected output to contain '[dry-run]', got: %s", out)
	}
	if _, statErr := os.Stat(filepath.Join(targetDir, "SKILL.md")); statErr == nil {
		t.Error("expected no SKILL.md after dry run")
	}
}

func TestSkillUnknownSubcommand(t *testing.T) {
	err := handleSkillCommand([]string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected error to mention 'unknown', got: %v", err)
	}
}

func TestSkillContentEmbedded(t *testing.T) {
	if skillContent == "" {
		t.Error("skillContent should not be empty")
	}
	if !strings.Contains(skillContent, "name: llm-mock") {
		t.Error("skillContent should contain 'name: llm-mock'")
	}
}