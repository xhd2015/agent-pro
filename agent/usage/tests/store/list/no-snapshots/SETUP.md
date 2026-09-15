# Scenario

**Feature**: a root that holds unrelated files and empty directories has no snapshots

```
usages/grok/2026-09-15/notes.txt     <- not a snapshot file
usages/providers.json                <- not a provider directory
List(0) -> no records, no error
```

## Preconditions

- The root exists but nothing in it ends with `-snapshot.jsonl`, so listing has to
  ignore what it does not recognize rather than fail on it.
- This is the state a fresh home reaches before the first collect, or after the
  user drops scratch files into the tree.

## Steps

1. Build an unrecognized tree under the root.
2. List it.

```go
import (
"os"
"path/filepath"
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
dir := filepath.Join(req.Root, "grok", "2026-09-15")
if err := os.MkdirAll(dir, 0o755); err != nil {
return err
}
if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello\n"), 0o644); err != nil {
return err
}
if err := os.WriteFile(filepath.Join(req.Root, "providers.json"), []byte("{}\n"), 0o644); err != nil {
return err
}
req.ListLasts = []int{0}
return nil
}
```
