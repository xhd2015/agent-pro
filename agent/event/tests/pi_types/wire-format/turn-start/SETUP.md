# Scenario

**Feature**: turn_start has no payload fields beyond `type`

## Preconditions
- turn_start has no payload fields beyond `type`.

## Steps
1. Parse the turn_start event JSON.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"turn_start"}`
	return nil
}
```
