# Scenario

**Feature**: empty multi list formats a friendly message

```
FormatPromptsListText(nil) -> contains "No user prompts found" + trailing \n
```

## Preconditions

- Op `format-empty` (does not call ListPrompts).
- No sessions required.

## Steps

1. Call FormatPromptsListText with nil/empty list.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-empty"
	req.FormatEmptyList = true
	return nil
}
```
