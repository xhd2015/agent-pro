# Scenario

**Feature**: empty Model + Effort stay empty (no agent-pro defaults)

```
Opts.Model="", ModelReasoningEffort=""
  -> capture both empty
  -> not gpt-5.6-luna / not max
```

## Steps

1. Leave both empty.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-lib-empty"
	req.Model = ""
	req.ModelReasoningEffort = ""
	return nil
}
```
