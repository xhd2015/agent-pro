# Scenario

**Feature**: Mode A follow-up shell-quotes an executable path that contains spaces

```
Executable = {temp}/My Tools/grok-fork
OpenInNewTerminal follow-up uses ShellQuote(Executable)
```

## Steps

1. Set `Executable` to a path with spaces.
2. Bare args.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Executable = filepath.Join(req.TempDir, "My Tools", "grok-fork")
	req.Args = []string{}
	return nil
}
```
