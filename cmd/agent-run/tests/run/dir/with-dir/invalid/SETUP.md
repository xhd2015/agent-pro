# Scenario

**Feature**: invalid `--dir` paths fail with a clear error

```
--dir <missing>     -> non-zero; stderr mentions missing / not found / does not exist
--dir <file-not-dir> -> non-zero; stderr mentions not a directory
```

## Steps

1. Leaves construct missing or file paths under TempDir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Invalid --dir branch: leaves supply missing or non-directory paths.
	t.Helper()
	if req.TempDir == "" {
		t.Fatal("TempDir must be set by root Setup before invalid --dir leaves")
	}
	return nil
}
```
