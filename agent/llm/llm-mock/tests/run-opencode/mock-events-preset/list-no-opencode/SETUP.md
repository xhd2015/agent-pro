# Scenario

**Feature**: `llm-mock run --mock-events-preset=list` exits 0 without opencode or mock server

```
llm-mock run --mock-events-preset=list -> stdout catalog -> exit 0
(no OPENCODE_CONFIG_DIR=, no mock listener)
```

## Steps

1. Set `ListOnly` and `MockEventsPreset=list`.
2. Do not set fake opencode hook (nothing should run).

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