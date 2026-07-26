# Scenario

**Feature**: explicit `--session ID` sets storage id; on TTY also registry id

```
agent-run run --session my-task "prompt"
  -> sessions/<runner>/my-task/
  -> [TTY] stderr + registry use my-task (same-id policy)
```

## Preconditions

- `--session-id-from-prompt` is not set.
- Session id uses registry-safe charset `[a-zA-Z0-9][a-zA-Z0-9._-]*`.

## Steps

1. Grouping records explicit-session mode on the request.
2. Children split non-TTY vs TTY and finalize `--session` + prompt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Mark mode for leaf clarity; leaves append --session and runner flags.
	if req.Prompt == "" {
		req.Prompt = "explicit session prompt"
	}
	return nil
}
```
