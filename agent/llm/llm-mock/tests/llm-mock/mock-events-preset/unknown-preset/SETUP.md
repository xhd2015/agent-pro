# Scenario

**Feature**: unknown `--mock-events-preset` name fails before server listens

```
llm-mock --mock-events-preset=nonexistent -> startup error (no listener)
```

## Steps

1. Set `MockEventsPreset` to a name not in the MVP catalog.
2. Do not send HTTP requests.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MockEventsPreset = "nonexistent"
	return nil
}
```