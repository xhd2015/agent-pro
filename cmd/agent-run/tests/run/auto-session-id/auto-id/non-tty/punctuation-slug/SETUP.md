# Scenario

**Feature**: punctuation in prompt is slugified to hyphens

```
agent-run run --auto-session-id "Hello, World!!"
  -> base slug hello-world (+ timestamp)
```

## Steps

1. Run with prompt `Hello, World!!`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "Hello, World!!"
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
