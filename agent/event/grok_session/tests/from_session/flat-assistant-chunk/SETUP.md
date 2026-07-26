# Scenario

**Feature**: flat agent_message_chunk converts to assistant ActionMessage

```
agent_message_chunk{"text":"Here is the answer"} -> ActionMessage role=assistant
```

## Preconditions
- One flat `agent_message_chunk` line.

## Steps
1. Provide one `acpAssistantChunk` wire line.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{acpAssistantChunk("Here is the answer")}
	return nil
}
```
