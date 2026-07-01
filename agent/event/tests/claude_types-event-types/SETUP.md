# Scenario

**Feature**: claude_types exposes event-type constants and a StreamEvent envelope that round-trips through JSON

```
# constants identify each top-level stream-json line type
claude_types.EventSystem / EventAssistant / EventUser / EventResult -> "system" / "assistant" / "user" / "result"

# a StreamEvent marshals back to the discriminated stream-json shape
StreamEvent{Type, Subtype, Message, Result, IsError, SessionID} -> JSON -> unmarshal -> equal
```

## Preconditions
- The `claude_types` package defines `EventType`, the four event-type constants (`EventSystem`, `EventAssistant`, `EventUser`, `EventResult`), and the `StreamEvent`, `Message`, and `ContentBlock` structs.

## Steps
1. Print the four event-type constant values.
2. Build a `StreamEvent` carrying an assistant `Message` with one `text` content block, marshal it to JSON, and append the JSON to the output.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	claude_types "github.com/xhd2015/agent-pro/agent/event/claude_types"
)

func Setup(t *testing.T, req *Request) error {
	var sb strings.Builder

	fmt.Fprintf(&sb, "EventSystem=%s\n", claude_types.EventSystem)
	fmt.Fprintf(&sb, "EventAssistant=%s\n", claude_types.EventAssistant)
	fmt.Fprintf(&sb, "EventUser=%s\n", claude_types.EventUser)
	fmt.Fprintf(&sb, "EventResult=%s\n", claude_types.EventResult)

	evt := claude_types.StreamEvent{
		Type:      claude_types.EventAssistant,
		SessionID: "sess_claude",
		Message: &claude_types.Message{
			Role: "assistant",
			Content: []claude_types.ContentBlock{
				{Type: "text", Text: "pong"},
			},
		},
	}
	data, _ := json.Marshal(evt)
	fmt.Fprintln(&sb, string(data))

	req.Output = sb.String()
	return nil
}
```
