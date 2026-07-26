# Scenario

**Feature**: single tool_call breakpoint encodes to one chat `tool_calls` entry (regression guard)

```
genQueue [tool_call bash]
POST chat -> tool_calls bash, finish_reason tool_calls
```

## Steps

1. `--mock-events-preset=tool-bash`.
2. Send one chat completion request.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.MockEventsPreset = "tool-bash"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"chat-tool-bash-guard"}]}`,
	}
	return nil
}
```