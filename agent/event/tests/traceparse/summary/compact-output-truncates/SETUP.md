# Scenario

**Feature**: traceparse leaf `summary/compact-output-truncates`

```
trace JSONL -> adapter registry -> parsed event JSON
```

## Preconditions
- Mode and inputs are set for this leaf.

## Steps
1. Configure `Request` fields for this scenario.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SubMode = "compact"
	b, err := os.ReadFile(filepath.Join(d.DOCTEST_CASE, "long-output.txt"))
	if err != nil {
		t.Fatal(err)
	}
	req.LongOutput = string(b)
	return nil
}
```
