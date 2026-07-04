# Scenario

**Feature**: agent-pro grok sessions list and session info from synthetic GROK_HOME

```
# harness builds synthetic Grok home with summary.json and optional signals.json
test harness -> sessions package -> fixtures under GROK_HOME/sessions/

# list discovers sessions; table shows MSGS and relative last-active times
sessions package -> FormatListTable(now) -> formatted terminal output

# info locates one session by exact UUID and renders detail blocks
sessions.Find/Info -> FormatInfoText(now) -> key-value session detail
```

## Preconditions

- Package `agent/grok/sessions` exposes List, FormatListTable, Find, Info, and
  FormatInfoText.
- Tests never read the real user `~/.grok` directory.

## Steps

1. Root Setup allocates `req.GrokHome` as `{temp}/.grok`.
2. Root Setup sets `req.Now` to a fixed UTC instant for deterministic relative times.
3. Leaf Setup writes `summary.json` fixtures under encoded cwd directories.

## Context

- Session path pattern:
  `{GrokHome}/sessions/{url.PathEscape(abs_cwd)}/{uuid}/summary.json`
- `last_active_at` drives sort order; `generated_title` populates the TITLE column.
- `num_chat_messages` populates the MSGS column in list table output.
- Info tests require exact session UUIDs (no prefix matching).

```go
import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const fixedNow = "2026-07-03T15:00:00.000Z"

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = filepath.Join(t.TempDir(), ".grok")
	now, err := time.Parse(time.RFC3339, fixedNow)
	if err != nil {
		t.Fatalf("parse fixed now: %v", err)
	}
	req.Now = now.UTC()
	return nil
}

type grokSessionOpts struct {
	NumMessages      int
	NumChatMessages  int
	CreatedAt        string
	UpdatedAt        string
	CurrentModelID   string
	AgentName        string
	SandboxProfile   string
	GitRootDir       string
	HeadBranch       string
	HeadCommit       string
	WriteUpdates     bool
	WriteSignals     bool
	WritePromptCtx   bool
	ContextTokens    int
	ContextWindow    int
	ContextUsagePct  int
	TokensBeforeComp int
}

func writeGrokSession(t *testing.T, grokHome, id, lastActiveAt, cwd, title string) string {
	t.Helper()
	return writeGrokSessionOpts(t, grokHome, id, lastActiveAt, cwd, title, grokSessionOpts{})
}

func writeGrokSessionOpts(t *testing.T, grokHome, id, lastActiveAt, cwd, title string, opts grokSessionOpts) string {
	t.Helper()
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatalf("abs cwd %q: %v", cwd, err)
	}
	encoded := url.PathEscape(absCwd)
	dir := filepath.Join(grokHome, "sessions", encoded, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}

	createdAt := opts.CreatedAt
	if createdAt == "" {
		createdAt = lastActiveAt
	}
	updatedAt := opts.UpdatedAt
	if updatedAt == "" {
		updatedAt = lastActiveAt
	}

	summary := map[string]any{
		"info": map[string]any{
			"id":  id,
			"cwd": absCwd,
		},
		"generated_title":     title,
		"created_at":          createdAt,
		"updated_at":          updatedAt,
		"last_active_at":      lastActiveAt,
		"num_messages":        opts.NumMessages,
		"num_chat_messages":   opts.NumChatMessages,
		"current_model_id":    opts.CurrentModelID,
		"agent_name":          opts.AgentName,
		"sandbox_profile":     opts.SandboxProfile,
		"git_root_dir":        opts.GitRootDir,
		"head_branch":         opts.HeadBranch,
		"head_commit":         opts.HeadCommit,
	}
	b, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	path := filepath.Join(dir, "summary.json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write summary %s: %v", path, err)
	}

	if opts.WriteUpdates {
		updatesPath := filepath.Join(dir, "updates.jsonl")
		if err := os.WriteFile(updatesPath, []byte(`{"type":"init"}`+"\n"), 0644); err != nil {
			t.Fatalf("write updates %s: %v", updatesPath, err)
		}
	}
	if opts.WriteSignals {
		signals := map[string]any{
			"contextTokensUsed":           opts.ContextTokens,
			"contextWindowTokens":         opts.ContextWindow,
			"contextWindowUsage":          opts.ContextUsagePct,
			"totalTokensBeforeCompaction": opts.TokensBeforeComp,
		}
		signalsPath := filepath.Join(dir, "signals.json")
		sb, err := json.Marshal(signals)
		if err != nil {
			t.Fatalf("marshal signals: %v", err)
		}
		if err := os.WriteFile(signalsPath, sb, 0644); err != nil {
			t.Fatalf("write signals %s: %v", signalsPath, err)
		}
	}
	if opts.WritePromptCtx {
		ctxPath := filepath.Join(dir, "prompt_context.json")
		if err := os.WriteFile(ctxPath, []byte(`{"bootstrap":true}`), 0644); err != nil {
			t.Fatalf("write prompt_context %s: %v", ctxPath, err)
		}
	}
	return path
}

func writeRawSummaryFile(t *testing.T, grokHome, encodedCwd, sessionID, content string) string {
	t.Helper()
	dir := filepath.Join(grokHome, "sessions", encodedCwd, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}
	path := filepath.Join(dir, "summary.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write raw summary %s: %v", path, err)
	}
	return path
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("operation failed: %v", resp.Err)
	}
}

func assertError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err == nil {
		t.Fatal("expected error but got nil")
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}
```