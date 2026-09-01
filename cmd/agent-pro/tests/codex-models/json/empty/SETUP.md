# Scenario

**Feature**: JSON empty catalog when home has no model files

```
# empty home -> Catalog with models=[]
List -> FormatJSON
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
