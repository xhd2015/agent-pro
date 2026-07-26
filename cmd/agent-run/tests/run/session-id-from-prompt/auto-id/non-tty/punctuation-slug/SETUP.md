# Scenario

**Feature**: punctuation in prompt is slugified to hyphens

```
agent-run run --session-id-from-prompt "Hello, World!!"
  -> base slug hello-world (+ timestamp)
```

## Steps

1. Run with prompt `Hello, World!!`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "Hello, World!!"
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
