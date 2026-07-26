# Scenario

**Feature**: `llm-mock run --mock-events-preset=list` exits 0 without grok or mock server

```
llm-mock run --mock-events-preset=list -> stdout catalog -> exit 0
(no GROK_HOME=, no mock listener)
```

## Steps

1. Set `ListOnly` and `MockEventsPreset=list`.
2. Do not set fake grok hook (nothing should run).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ListOnly = true
	req.MockEventsPreset = "list"
	req.ExpectedExit = 0
	return nil
}
```