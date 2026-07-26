# Scenario

**Feature**: absolute `--dir` path from a different process cwd

```
cwd=TempDir; --dir <abs workspace under TempDir>
  -> meta.workspace = abs workspace ≠ process cwd
```

## Steps

1. Leaves create an absolute workspace dir distinct from process cwd.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Absolute-path branch under with-dir; leaves create abs workspace fixtures.
	t.Helper()
	if req.TempDir == "" {
		t.Fatal("TempDir must be set by root Setup before absolute --dir leaves")
	}
	return nil
}
```
