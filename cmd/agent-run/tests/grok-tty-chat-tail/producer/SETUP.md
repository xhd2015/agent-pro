# Scenario

**Bug**: producer grok tail must keep appending after initial `updates.jsonl` sync under keep-tty

```
keep-tty run + partial updates.jsonl (user, think, pending tool_call)
  -> delayed tool_call_update + agent_message_chunk + turn_completed
  -> events.jsonl is the assertion surface
```

## Steps

1. Grouping setup sets `req.Mode = "producer"` and keep-tty producer defaults.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "producer"
	req.ExecTimeout = producerFinishTimeout
	return nil
}
```