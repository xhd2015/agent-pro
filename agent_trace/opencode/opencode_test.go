package opencode

import (
	"strings"
	"testing"
)

func TestOpencodeToolActivityBashShowsCommand(t *testing.T) {
	part := &opencodeRunPart{
		Type: "tool",
		Tool: "bash",
		State: &opencodePartState{
			Status: "completed",
			Title:  "Run Go tests to verify changes",
			Output: "ok  \texample.com/hello-world\t1.717s\n",
			Input: map[string]any{
				"command":     "go test ./...",
				"description": "Run Go tests to verify changes",
				"workdir":     "/tmp/hello-world",
			},
		},
	}

	activity := opencodeToolActivity(part)
	if activity == nil {
		t.Fatal("expected activity, got nil")
	}

	if !strings.Contains(activity.Summary, "go test ./...") {
		t.Errorf("summary should contain the command 'go test ./...', got:\n%s", activity.Summary)
	}
	if !strings.Contains(activity.Summary, "Run Go tests to verify changes") {
		t.Errorf("summary should contain the title 'Run Go tests to verify changes', got:\n%s", activity.Summary)
	}
	if !strings.Contains(activity.Summary, "ok") {
		t.Errorf("summary should contain output, got:\n%s", activity.Summary)
	}
	if activity.ToolName != "Shell" {
		t.Errorf("tool name = %q, want Shell", activity.ToolName)
	}
	if activity.Status != "completed" {
		t.Errorf("status = %q, want completed", activity.Status)
	}
}

func TestOpencodeToolActivityBashCommandOnly(t *testing.T) {
	part := &opencodeRunPart{
		Type: "tool",
		Tool: "bash",
		State: &opencodePartState{
			Status: "running",
			Input: map[string]any{
				"command": "echo hello",
			},
		},
	}

	activity := opencodeToolActivity(part)
	if activity == nil {
		t.Fatal("expected activity, got nil")
	}

	expected := "echo hello"
	if activity.Summary != expected {
		t.Errorf("summary = %q, want %q", activity.Summary, expected)
	}
}

func TestOpencodeToolActivityRead(t *testing.T) {
	part := &opencodeRunPart{
		Type: "tool",
		Tool: "read",
		State: &opencodePartState{
			Status: "completed",
			Title:  "path/to/main.go",
			Output: "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}",
			Input: map[string]any{
				"filePath": "/path/to/main.go",
			},
		},
	}

	activity := opencodeToolActivity(part)
	if activity == nil {
		t.Fatal("expected activity, got nil")
	}

	if !strings.Contains(activity.Summary, "path/to/main.go") {
		t.Errorf("summary should contain the file path, got:\n%s", activity.Summary)
	}
	if !strings.Contains(activity.Summary, "package main") {
		t.Errorf("summary should contain file content, got:\n%s", activity.Summary)
	}
	if activity.ToolName != "Read File" {
		t.Errorf("tool name = %q, want Read File", activity.ToolName)
	}
}

func TestOpencodeToolActivityError(t *testing.T) {
	part := &opencodeRunPart{
		Type: "tool",
		Tool: "bash",
		State: &opencodePartState{
			Status: "error",
			Title:  "Run failing command",
			Output: "some output",
			Error:  "command not found",
			Input: map[string]any{
				"command": "nonexistent-command",
			},
		},
	}

	activity := opencodeToolActivity(part)
	if activity == nil {
		t.Fatal("expected activity, got nil")
	}

	if activity.Status != "failed" {
		t.Errorf("status = %q, want failed", activity.Status)
	}
	if activity.Summary != "command not found" {
		t.Errorf("summary = %q, want 'command not found'", activity.Summary)
	}
}

func TestOpencodeToolActivityNoState(t *testing.T) {
	part := &opencodeRunPart{
		Type: "tool",
		Tool: "unknown-tool",
	}

	activity := opencodeToolActivity(part)
	if activity == nil {
		t.Fatal("expected activity, got nil")
	}

	if activity.ToolName != "unknown-tool" {
		t.Errorf("tool name = %q, want unknown-tool", activity.ToolName)
	}
	if activity.Summary != "" {
		t.Errorf("summary = %q, want empty", activity.Summary)
	}
}

func TestOpencodeToolActivityFallbackInput(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantSum string
	}{
		{
			name: "path fallback",
			input: map[string]any{
				"path": "/some/file.go",
			},
			wantSum: "/some/file.go",
		},
		{
			name: "pattern fallback",
			input: map[string]any{
				"pattern": "**/*.go",
			},
			wantSum: "**/*.go",
		},
		{
			name: "query fallback",
			input: map[string]any{
				"query": "search term",
			},
			wantSum: "search term",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := &opencodeRunPart{
				Type: "tool",
				Tool: "search",
				State: &opencodePartState{
					Status: "completed",
					Input:  tt.input,
				},
			}

			activity := opencodeToolActivity(part)
			if activity == nil {
				t.Fatal("expected activity, got nil")
			}

			if activity.Summary != tt.wantSum {
				t.Errorf("summary = %q, want %q", activity.Summary, tt.wantSum)
			}
		})
	}
}
