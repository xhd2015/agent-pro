# Scenario

**Feature**: `llm-mock run --mock-events-preset=list` exits 0 without codex or mock server

```
llm-mock run --mock-events-preset=list -> stdout catalog -> exit 0
(no CODEX_HOME=, no mock listener)
```

## Steps

1. Set `ListOnly` and `MockEventsPreset=list`.
2. Do not set fake codex hook (nothing should run).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ListOnly = true
	req.MockEventsPreset = "list"
	req.ExpectedExit = 0
	return nil
}
```