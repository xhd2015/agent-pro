# Scenario

**Feature**: grok think+message split must not duplicate message in agent-events log

Reproduces `llm-mock run --log-events test.jsonl grok` + one user turn where grok issues
two chat completion HTTP calls (think response, then message response).

Bug: first request logs think **and** peeked message; second request logs message again → 3 lines.

```
POST #1 user:Hello -> think served -> agent-events: think + message (peek) WRONG
POST #2 user:Hello -> message served -> agent-events: message again DUPLICATE
```

## Steps

1. Empty `exchanges[]` (random fallback).
2. Two identical single-message requests (simulates grok per-event HTTP split).
3. Read `--agent-events-file` after both requests.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": []
}`
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"Hello"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"Hello"}]}`,
	}
	return nil
}
```