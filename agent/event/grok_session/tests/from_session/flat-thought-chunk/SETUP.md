# Scenario

**Feature**: flat agent_thought_chunk converts to ActionThink with turn_index=0

```
agent_thought_chunk{"text":"planning"} -> ActionThink turn_index=0
```

## Preconditions
- One flat `agent_thought_chunk` line.

## Steps
1. Provide one `acpThoughtChunk("planning ls output")` wire line.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.WireLines = []string{acpThoughtChunk("planning ls output")}
	return nil
}
```
