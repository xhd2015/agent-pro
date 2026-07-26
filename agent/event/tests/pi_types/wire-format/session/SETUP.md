# Scenario

**Feature**: A session event JSON has `type` and `id` fields

## Preconditions
- A session event JSON has `type` and `id` fields.

## Steps
1. Parse the session event JSON into a `pi_types.Event`.
2. Marshal it back to JSON.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"session","id":"sess_abc123"}`
	return nil
}
```
