# Scenario

**Feature**: thinking blocks coalesce consecutive agent_thought_chunk runs

```
# updates.jsonl with flat + nested thought chunks and a non-thought gap
writeGrokSessionOpts -> writeUpdatesJSONL -> sessions.Stats

# consecutive chunks = 1 block; gap starts a new block; nested envelope counts
```

## Preconditions

- Session has summary + updates only (no signals/events required for thinking).
- Sequence:
  1. flat thought "a"
  2. flat thought "b"  → still run #1
  3. nested thought "c" → still run #1 (consecutive thoughts)
  4. non-thought `agent_message_chunk` → ends run #1
  5. flat thought "d" → starts run #2
  6. nested thought "e" → still run #2
- Expected `ThinkingBlocks` = 2.

## Steps

1. Write a minimal session summary.
2. Write `updates.jsonl` with the thought sequence above.
3. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const thinkingBlocksSessionID = "019f283b-dddd-7ddd-dddd-dddddddddddd"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = thinkingBlocksSessionID
	summaryPath := writeGrokSessionOpts(t, req.GrokHome, thinkingBlocksSessionID,
		"2026-07-03T14:20:00.000Z",
		"/tmp/grok-stats-thinking",
		"Thinking blocks session",
		grokSessionOpts{
			NumMessages:     4,
			NumChatMessages: 2,
		})
	writeUpdatesJSONL(t, sessionDirOf(summaryPath), []map[string]any{
		thoughtChunk("a"),
		thoughtChunk("b"),
		nestedThoughtChunk("c"),
		nonThoughtUpdate("agent_message_chunk"),
		thoughtChunk("d"),
		nestedThoughtChunk("e"),
	})
	return nil
}
```
