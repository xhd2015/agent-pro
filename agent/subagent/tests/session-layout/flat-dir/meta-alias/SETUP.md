# Scenario

**Feature**: agent_session_id alias resolves flat sessions without explicit_session_id

```
# meta.json has agent_session_id only
subagent.Run matches alias -> resumes/completes without rewriting host schema
```

## Preconditions

- Pre-created `meta.json` uses `agent_session_id` and omits `explicit_session_id`.

## Steps

1. Descendant leaves write host meta and run with matching `SessionID`.

## Context

- Distinct from `merged-meta` which focuses on field preservation after opencode update.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.AgentRunner == "" {
		req.AgentRunner = "fake-codex"
	}
	return nil
}```
