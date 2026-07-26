# Scenario

**Feature**: envelope wire line parses same as flat user chunk

```
{"method":"_x.ai/session/update",...} -> same ActionMessage as flat user_message_chunk
```

## Preconditions
- Envelope wraps a `user_message_chunk` update.

## Steps
1. Provide `acpEnvelopeUserChunk` with text `run ls`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{acpEnvelopeUserChunk(req.SessionID, "run ls")}
	return nil
}
```
