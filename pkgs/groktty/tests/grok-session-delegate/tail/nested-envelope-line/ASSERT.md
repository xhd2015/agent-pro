## Expected

- Emits user `ActionMessage` with text `run ls`.
- Emits assistant `ActionMessage` with text `listing files`.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	foundUser := false
	foundAssistant := false
	for _, ev := range resp.Events {
		if ev.Type != types.ActionMessage {
			continue
		}
		switch ev.Role {
		case "user":
			if ev.Text == "run ls" {
				foundUser = true
			}
		case "assistant":
			if ev.Text == "listing files" {
				foundAssistant = true
			}
		}
	}
	if !foundUser {
		t.Fatalf("missing user message run ls in:\n%s", formatEvents(resp.Events))
	}
	if !foundAssistant {
		t.Fatalf("missing assistant message listing files in:\n%s", formatEvents(resp.Events))
	}
}
```