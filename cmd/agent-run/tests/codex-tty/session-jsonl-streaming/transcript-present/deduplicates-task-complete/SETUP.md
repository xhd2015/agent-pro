# Scenario

**Feature**: duplicate `task_complete.last_agent_message` is not emitted twice

```
rollout JSONL event_msg.agent_message "final"
  -> later event_msg.task_complete.last_agent_message "final"
  -> stdout contains one assistant final answer
```

## Preconditions

- The duplicate text is byte-for-byte identical.

## Steps

1. Seed `agent_message` followed by duplicate `task_complete`.
2. Assert stdout contains the final answer exactly once.

```go
import "testing"

const codexDedupedFinalText = "JSONL_DEDUPED_FINAL_MESSAGE"

func Setup(t *testing.T, req *Request) error {
	seedCodexTranscript(t, req,
		codexAgentMessageLine(codexDedupedFinalText),
		codexTaskCompleteLine(codexDedupedFinalText),
	)
	req.StreamProbeSubstring = codexDedupedFinalText
	return nil
}
```
