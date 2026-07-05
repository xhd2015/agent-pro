# Scenario

**Feature**: nested envelope wire line tails correctly

```
envelope user_message_chunk + flat assistant + turn_completed -> user and assistant text emitted
```

## Preconditions

- User chunk uses `_x.ai/session/update` envelope wire format.
- Assistant chunk is flat ACP format.

## Steps

1. Provide envelope user chunk and flat assistant chunk ending in `turn_completed`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WireLines = []string{
		acpEnvelopeUserChunk(req.SessionID, "run ls"),
		acpAssistantChunk("listing files"),
		acpTurnCompleted(),
	}
	return nil
}
```