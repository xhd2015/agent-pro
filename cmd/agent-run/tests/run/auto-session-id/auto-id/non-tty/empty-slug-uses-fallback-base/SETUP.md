# Scenario

**Feature**: when slugify yields empty base, fall back to `sess` or `task`

```
agent-run run --auto-session-id "!!!"
  -> base sess|task + -YYYYMMDD-HHMMSS
```

## Steps

1. Run with punctuation-only prompt `!!!`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "!!!"
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
