# Scenario

**Feature**: detailed session info for a single Grok session

```
# locate session dir by exact UUID under GROK_HOME/sessions
sessions.Find -> sessions.Info(grokHome, sessionID)

# render summary fields, file paths, optional token usage
SessionInfo -> FormatInfoText(now) -> key-value output
```

## Preconditions

- This branch tests the `info` operation.
- Session ID must be a full UUID match (no prefix matching).

## Steps

1. Set `req.Operation = "info"`.
2. Leaf Setup writes `summary.json` and optional sidecar files for `req.SessionID`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "info"
	return nil
}
```