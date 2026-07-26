# Scenario

**Feature**: analyse one Grok session into counts, latency, tool times, and task aggregates

```
# locate session by exact UUID under GROK_HOME/sessions
sessions.Find -> sessions.Stats(grokHome, sessionID)

# prefer signals for counts/latency; events for tool handler time; updates for
# thinking blocks, background tasks, and subagents
SessionStats -> FormatStatsTextOpts(stats, {Home, Now, ColorMode, TopN})
  -> sectioned text (pretty durations, sorted tool table, optional Top-N, color)
  -> Top background: # DURATION EXIT COMMAND (cmd display ≤200 runes + …)
  -> Top subagents:  # DURATION STATUS TYPE TOOLS TURNS DESC (desc ≤64 runes)
```

## Preconditions

- This branch tests the `stats` operation.
- Session ID must be a full UUID match (no prefix matching).
- Missing optional sidecars (signals/events/updates) warn and still succeed when
  `summary.json` exists.
- Format leaves set `ColorMode` / `TopN`+`TopNSet` when they care about
  presentation; aggregation leaves leave them unset (Run: never color, TopN=5).
- Rich top-item model (replaces `TimedItem` for bg/sub lists):
  - `BackgroundTaskItem`: DurationMs, Command (full in Stats), Description,
    ExitCode (*int), Kind, CWD
  - `SubagentItem`: DurationMs, ID, Description, Type, Status, ToolCalls,
    Turns, TokensUsed, Model

## Steps

1. Set `req.Operation = "stats"`.
2. Leaf Setup writes `summary.json` and optional `signals.json`, `events.jsonl`,
   and `updates.jsonl` for `req.SessionID`.

## Context

- Stats identity comes from summary via Find.
- Counts/latency prefer signals.json camelCase keys.
- Tool handler times come from events.jsonl `tool_completed.duration_ms`.
- Thinking blocks coalesce consecutive `agent_thought_chunk` updates.
- Background tasks:
  - optional `task_backgrounded` maps `task_id` → description
  - `task_completed` wall clock from start/end; full `task_snapshot.command`
    (no store truncate); optional `exit_code`; join description by task_id
  - Top table: `#  DURATION  EXIT  COMMAND` — EXIT is int or `-` if nil;
    COMMAND display-truncated at **200 runes** then `…`
- Subagents:
  - `subagent_spawned` maps id → description, type, model
  - `subagent_finished` duration_ms / status / tool_calls / turns / tokens_used;
    join spawn meta by `subagent_id`; DESC display-truncate **64 runes**;
    empty desc falls back to short ID
  - Top table: `#  DURATION  STATUS  TYPE  TOOLS  TURNS  DESC`
- Human formatter: pretty durations; tool table by N desc; Top-N by total time.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "stats"
	return nil
}

// thoughtChunk builds a flat agent_thought_chunk update line.
func thoughtChunk(text string) map[string]any {
	return map[string]any{
		"sessionUpdate": "agent_thought_chunk",
		"content": map[string]any{
			"type": "text",
			"text": text,
		},
	}
}

// nestedThoughtChunk builds a nested params.update thought chunk (wire envelope).
func nestedThoughtChunk(text string) map[string]any {
	return map[string]any{
		"params": map[string]any{
			"update": thoughtChunk(text),
		},
	}
}

// nonThoughtUpdate is any non-thought session update that ends a thought run.
func nonThoughtUpdate(kind string) map[string]any {
	return map[string]any{
		"sessionUpdate": kind,
	}
}

// toolCompleted builds an events.jsonl tool_completed line.
func toolCompleted(name string, durationMs int, outcome string) map[string]any {
	return map[string]any{
		"type":        "tool_completed",
		"tool_name":   name,
		"duration_ms": durationMs,
		"outcome":     outcome,
		"ts":          "2026-07-03T13:00:00.000Z",
	}
}

// toolStarted builds an events.jsonl tool_started line.
func toolStarted(name string) map[string]any {
	return map[string]any{
		"type":      "tool_started",
		"tool_name": name,
		"ts":        "2026-07-03T13:00:00.000Z",
	}
}

// taskCompleted builds an updates.jsonl task_completed line with wall-clock times.
func taskCompleted(startSec, endSec int64) map[string]any {
	return map[string]any{
		"sessionUpdate": "task_completed",
		"task_snapshot": map[string]any{
			"start_time": map[string]any{
				"secs_since_epoch":  startSec,
				"nanos_since_epoch": 0,
			},
			"end_time": map[string]any{
				"secs_since_epoch":  endSec,
				"nanos_since_epoch": 0,
			},
		},
	}
}

// taskCompletedCmd is taskCompleted plus a full command for Top background tasks.
func taskCompletedCmd(startSec, endSec int64, command string) map[string]any {
	m := taskCompleted(startSec, endSec)
	snap := m["task_snapshot"].(map[string]any)
	snap["command"] = command
	return m
}

// taskCompletedCmdExit is taskCompletedCmd plus exit_code on task_snapshot.
func taskCompletedCmdExit(startSec, endSec int64, command string, exitCode int) map[string]any {
	m := taskCompletedCmd(startSec, endSec, command)
	snap := m["task_snapshot"].(map[string]any)
	snap["exit_code"] = exitCode
	return m
}

// taskBackgrounded records optional description keyed by task_id (joined at complete).
func taskBackgrounded(taskID, description string) map[string]any {
	return map[string]any{
		"sessionUpdate": "task_backgrounded",
		"task_id":       taskID,
		"description":   description,
	}
}

// subagentFinished builds an updates.jsonl subagent_finished line.
func subagentFinished(durationMs int, status string) map[string]any {
	return map[string]any{
		"sessionUpdate": "subagent_finished",
		"duration_ms":   durationMs,
		"status":        status,
	}
}

// subagentFinishedDesc is subagentFinished plus a description (legacy / finish-only).
func subagentFinishedDesc(durationMs int, status, description string) map[string]any {
	m := subagentFinished(durationMs, status)
	m["description"] = description
	return m
}

// subagentSpawned maps subagent_id → description, type, model for join at finish.
func subagentSpawned(id, description, typ, model string) map[string]any {
	return map[string]any{
		"sessionUpdate": "subagent_spawned",
		"subagent_id":   id,
		"description":   description,
		"type":          typ,
		"model":         model,
	}
}

// subagentFinishedFull is finish with id + status/tools/turns/tokens (join spawn meta).
func subagentFinishedFull(id string, durationMs int, status string, toolCalls, turns, tokensUsed int) map[string]any {
	return map[string]any{
		"sessionUpdate": "subagent_finished",
		"subagent_id":   id,
		"duration_ms":   durationMs,
		"status":        status,
		"tool_calls":    toolCalls,
		"turns":         turns,
		"tokens_used":   tokensUsed,
	}
}

// sectionAfter returns text from the first occurrence of header through the next
// blank-line-led section or end of string. Used by format asserts for Top tables.
func sectionAfter(out, header string) string {
	i := strings.Index(out, header)
	if i < 0 {
		return ""
	}
	rest := out[i:]
	// Skip the header line itself for stop search, keep it in returned section.
	nl := strings.Index(rest, "\n")
	if nl < 0 {
		return rest
	}
	body := rest[nl+1:]
	// Stop at next major section title line (non-indented non-empty after blank).
	stops := []string{
		"\nTop tools",
		"\nTop background",
		"\nTop subagents",
		"\nBackground tasks",
		"\nSubagents",
		"\nSources",
		"\nTool handler",
		"\nCounts",
		"\nLatency",
	}
	cut := len(body)
	for _, s := range stops {
		// Avoid stopping on the same header we started with.
		if strings.HasPrefix(header, strings.TrimSpace(s)) {
			continue
		}
		if j := strings.Index(body, s); j >= 0 && j < cut {
			cut = j
		}
	}
	return rest[:nl+1+cut]
}
```