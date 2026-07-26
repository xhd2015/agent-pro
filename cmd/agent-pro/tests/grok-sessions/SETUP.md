# Scenario

**Feature**: agent-pro grok sessions list, grep-filter, session info, and session stats from synthetic GROK_HOME

```
# harness builds synthetic Grok home with summary.json, optional chat_history.jsonl,
# signals.json, events.jsonl, updates.jsonl
test harness -> sessions package -> fixtures under GROK_HOME/sessions/

# list discovers sessions; optional grep filters by content hits
sessions package -> List | ListWithGrep -> FormatListTable | FormatListTableWithHits

# info locates one session by exact UUID and renders detail blocks
sessions.Find/Info -> FormatInfoText(now) -> key-value session detail

# stats aggregates signals / events / updates into SessionStats
sessions.Stats(grokHome, sessionID) -> FormatStatsTextOpts(stats, opts)
# opts: Home, Now, ColorMode (never|always|auto), TopN (0 hides Top sections)
```

## Preconditions

- Package `agent/grok/sessions` exposes List, FormatListTable, ListWithGrep,
  FormatListTableWithHits, Find, Info, FormatInfoText, Stats,
  FormatStatsText, and FormatStatsTextOpts.
- Expected grep types:
  - `MatchHit` with File, Line, Part, Snippet, MatchStart, MatchLen
  - `SessionMatch` embedding Session plus `Hits []MatchHit`
  - `ListWithGrep(grokHome, limit, pattern) ([]SessionMatch, error)`
  - `FormatListTableWithHits(matches, home, now, colorMode string) string`
    where colorMode is `never` | `always` | `auto`
- Expected stats types:
  - `SessionStats` with identity, counts, latency, Tools, BackgroundTasks,
    Subagents, Sources, plus optional `BackgroundTaskItems` /
    `SubagentItems` for Top sections
  - `BackgroundTaskItem` (DurationMs, Command full in Stats, Description,
    ExitCode *int, Kind, CWD) and `SubagentItem` (DurationMs, ID, Description,
    Type, Status, ToolCalls, Turns, TokensUsed, Model) replace `TimedItem`
    for bg/sub lists (`TimedItem` may remain unused or for tools only)
  - `ToolStat`, `TaskAgg`, `StatsSources`
  - `Stats(grokHome, sessionID string) (*SessionStats, error)`
  - `FormatStatsOptions` with Home, Now, ColorMode, TopN
  - `FormatStatsTextOpts(stats *SessionStats, opts FormatStatsOptions) string`
    — preferred harness entry (pretty durations, tool table sort, color, tops)
  - Top background table: `#  DURATION  EXIT  COMMAND` (cmd display ≤**200**
    runes + `…`; EXIT integer or `-`)
  - Top subagents table: `#  DURATION  STATUS  TYPE  TOOLS  TURNS  DESC`
    (desc display ≤**64** runes; join spawn→finish by `subagent_id`)
  - `FormatStatsText(stats, home, now)` may remain as a thin wrapper
    (`ColorMode: "never"`, `TopN: 5`) for CLI / callers that do not pass opts
- Harness `Request` fields for stats formatting:
  - `ColorMode` empty → Run uses `"never"`
  - `TopNSet` false → Run uses TopN `5`; `TopNSet` true uses `TopN` (`0` hides tops)
- Tests never read the real user `~/.grok` directory.

## Steps

1. Root Setup allocates `req.GrokHome` as `{temp}/.grok`.
2. Root Setup sets `req.Now` to a fixed UTC instant for deterministic relative times.
3. Leaf Setup writes `summary.json` and optional sidecars (`chat_history.jsonl`,
   `signals.json`, `events.jsonl`, `updates.jsonl`) under encoded cwd directories.

## Context

- Session path pattern:
  `{GrokHome}/sessions/{url.PathEscape(abs_cwd)}/{uuid}/summary.json`
- `last_active_at` drives sort order; `generated_title` populates the TITLE column.
- `num_chat_messages` populates the MSGS column in list table output.
- Grep v1 searches `summary.json` + `chat_history.jsonl` only (not `updates.jsonl`).
- Info and stats tests require exact session UUIDs (no prefix matching).
- Stats prefers `signals.json` for session-level counts/latency; uses
  `events.jsonl` for per-tool handler durations; uses `updates.jsonl` for
  thinking-block coalescing, background-task wall clock, and subagent durations.
- Missing optional stats files warn via `Sources.Warnings` and do not fail.
- Human stats output uses pretty durations; tool table sorted by N desc;
  optional Top-N sections; ColorMode controls ANSI (tests default never).

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

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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
	SessionSummary   string
	WriteUpdates     bool
	WriteSignals     bool
	WritePromptCtx   bool
	ContextTokens    int
	ContextWindow    int
	ContextUsagePct  int
	TokensBeforeComp int
}

// chatHistoryMsg is a simplified chat_history.jsonl line for grep fixtures.
// Type mirrors Grok message types: system, user, assistant, reasoning, tool_result.
type chatHistoryMsg struct {
	Type    string
	Content string
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
		"generated_title":   title,
		"created_at":        createdAt,
		"updated_at":        updatedAt,
		"last_active_at":    lastActiveAt,
		"num_messages":      opts.NumMessages,
		"num_chat_messages": opts.NumChatMessages,
		"current_model_id":  opts.CurrentModelID,
		"agent_name":        opts.AgentName,
		"sandbox_profile":   opts.SandboxProfile,
		"git_root_dir":      opts.GitRootDir,
		"head_branch":       opts.HeadBranch,
		"head_commit":       opts.HeadCommit,
	}
	if opts.SessionSummary != "" {
		summary["session_summary"] = opts.SessionSummary
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

// writeChatHistory writes chat_history.jsonl next to summary.json.
// Formats mirror real Grok CLI files:
//   - user: content as [{type:text,text:...}] array
//   - assistant/system/tool_result: content as string
//   - reasoning: summary as [{type:summary_text,text:...}]
func writeChatHistory(t *testing.T, sessionDir string, msgs []chatHistoryMsg) string {
	t.Helper()
	var b strings.Builder
	for _, m := range msgs {
		var obj map[string]any
		switch m.Type {
		case "user":
			obj = map[string]any{
				"type": "user",
				"content": []map[string]any{
					{"type": "text", "text": m.Content},
				},
			}
		case "reasoning":
			obj = map[string]any{
				"type": "reasoning",
				"summary": []map[string]any{
					{"type": "summary_text", "text": m.Content},
				},
			}
		default:
			// system, assistant, tool_result, or any other type with string content
			obj = map[string]any{
				"type":    m.Type,
				"content": m.Content,
			}
		}
		line, err := json.Marshal(obj)
		if err != nil {
			t.Fatalf("marshal chat history line: %v", err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	path := filepath.Join(sessionDir, "chat_history.jsonl")
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		t.Fatalf("write chat history %s: %v", path, err)
	}
	return path
}

func sessionDirOf(summaryPath string) string {
	return filepath.Dir(summaryPath)
}

// writeSignalsJSON writes (or overwrites) signals.json in a session directory.
// Used by stats fixtures for richer counters beyond the token-only WriteSignals path.
func writeSignalsJSON(t *testing.T, sessionDir string, signals map[string]any) string {
	t.Helper()
	path := filepath.Join(sessionDir, "signals.json")
	b, err := json.Marshal(signals)
	if err != nil {
		t.Fatalf("marshal signals: %v", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write signals %s: %v", path, err)
	}
	return path
}

// writeJSONL writes path with one JSON object per line.
func writeJSONL(t *testing.T, path string, lines []map[string]any) string {
	t.Helper()
	var b strings.Builder
	for _, line := range lines {
		raw, err := json.Marshal(line)
		if err != nil {
			t.Fatalf("marshal jsonl line: %v", err)
		}
		b.Write(raw)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		t.Fatalf("write jsonl %s: %v", path, err)
	}
	return path
}

// writeEventsJSONL writes events.jsonl (tool_started / tool_completed, etc.).
func writeEventsJSONL(t *testing.T, sessionDir string, events []map[string]any) string {
	t.Helper()
	return writeJSONL(t, filepath.Join(sessionDir, "events.jsonl"), events)
}

// writeUpdatesJSONL writes updates.jsonl (thought chunks, tasks, subagents).
func writeUpdatesJSONL(t *testing.T, sessionDir string, updates []map[string]any) string {
	t.Helper()
	return writeJSONL(t, filepath.Join(sessionDir, "updates.jsonl"), updates)
}

// defaultStatsSignals returns a realistic signals.json map for stats fixtures.
func defaultStatsSignals() map[string]any {
	return map[string]any{
		"turnCount":               2,
		"userMessageCount":        2,
		"assistantMessageCount":   5,
		"toolCallCount":           3,
		"toolFailureCount":        1,
		"errorCount":              0,
		"cancellationCount":       0,
		"compactionCount":         1,
		"sessionDurationSeconds":  120,
		"avgResponseTimeMs":       1500,
		"avgTimeToFirstTokenMs":   400,
		"contextTokensUsed":       75085,
		"contextWindowTokens":     200000,
		"contextWindowUsage":      38,
		"totalTokensBeforeCompaction": 0,
	}
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

func assertFloatNear(t *testing.T, name string, got, want, eps float64) {
	t.Helper()
	if got < want-eps || got > want+eps {
		t.Fatalf("%s = %v, want ~%v (±%v)", name, got, want, eps)
	}
}

func joinWarnings(warnings []string) string {
	return strings.Join(warnings, "\n")
}
```
