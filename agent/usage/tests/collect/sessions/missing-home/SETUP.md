# Scenario

**Feature**: a provider the user never ran a session with reports no sessions instead of an error

```
GrokHome/CodexHome/CommandCodeHome -> fresh dirs with auth.json, no session tree
Collect -> SessionsBlock{total: 0}, no oldest, no newest, error empty
```

## Preconditions

- All three homes are fresh directories holding only `auth.json`, so the usage
  fetches still work but no session tree exists.
- The session walk must treat a missing tree as "no sessions" rather than as a
  failure, or every snapshot of a signed-in-but-idle provider would carry an error.

## Steps

1. Point every home at a directory that holds credentials and nothing else.

```go
import (
"os"
"path/filepath"
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
tmp := t.TempDir()
req.GrokHome = filepath.Join(tmp, "grok")
req.CodexHome = filepath.Join(tmp, "codex")
req.CommandCodeHome = filepath.Join(tmp, "commandcode")

for dir, auth := range map[string]string{
req.GrokHome:        grokAuth,
req.CodexHome:       codexAuth,
req.CommandCodeHome: commandCodeAuth,
} {
if err := os.MkdirAll(dir, 0o755); err != nil {
return err
}
if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(auth), 0o600); err != nil {
return err
}
}
return nil
}
```
