# Scenario

**Feature**: multi-chunk coalescing roundtrips semantically

```
multiple chunks of same kind coalesce forward; reverse may re-chunk; semantics equal
```

## Preconditions
- Multiple assistant chunks coalesce on forward pass.

## Steps
1. Seed split assistant chunks within one turn.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WireLines = []string{
		acpUserChunk("prompt"),
		acpAssistantChunk("Hello "),
		acpAssistantChunk("world"),
		acpTurnCompleted(),
	}
	return nil
}
```
