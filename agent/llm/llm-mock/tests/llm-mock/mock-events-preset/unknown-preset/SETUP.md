# Scenario

**Feature**: unknown `--mock-events-preset` name fails before server listens

```
llm-mock --mock-events-preset=nonexistent -> startup error (no listener)
```

## Steps

1. Set `MockEventsPreset` to a name not in the MVP catalog.
2. Do not send HTTP requests.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.MockEventsPreset = "nonexistent"
	return nil
}
```