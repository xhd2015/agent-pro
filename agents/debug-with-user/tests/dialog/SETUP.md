# Scenario

**Feature**: dialog string utilities for AppleScript integration

```
title/message/options -> Escape (per field) -> osascript script fragments
osascript completion -> ParseOsascriptOutput -> structured answer fields
```

## Preconditions

- `dialog` package is importable from module root.

## Steps

1. Grouping setup pins tests under the `dialog` operation family.
2. Leaf setup sets `req.Operation` to `escape` or `parse`.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	switch req.Operation {
	case "", "escape", "parse":
		return nil
	default:
		return fmt.Errorf("dialog subtree supports escape|parse, got %q", req.Operation)
	}
}
```
