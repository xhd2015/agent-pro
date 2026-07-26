# Scenario

**Feature**: turn_start has no payload fields beyond `type`

## Preconditions
- turn_start has no payload fields beyond `type`.

## Steps
1. Parse the turn_start event JSON.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"turn_start"}`
	return nil
}
```
