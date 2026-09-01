# Scenario

**Feature**: empty Codex home yields friendly no-models message

```
# missing config/cache soft-fail to empty Catalog
List(emptyHome) -> Models=[]

# FormatText prints "(no models)"
```

## Preconditions

- Empty Codex home directory.

## Steps

1. Ensure home directory exists; write no fixtures.

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := os.MkdirAll(req.CodexHome, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return nil
}
```
