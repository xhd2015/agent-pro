# Scenario

**Feature**: new sessions stamp `meta.workspace` from selected workspace

```
# create session after PUT selected
PUT selected -> POST /sessions {runner, prompt}
  -> response session.workspace == selected
```

## Preconditions

- Valid selected directory.
- Runner `fake-codex` (async start; response returns immediately).

## Steps

1. Leaves PUT then POST sessions.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Scenario == "" {
		req.Scenario = "sessions"
	}
	return nil
}
```
