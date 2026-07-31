# Scenario

**Feature**: focus help documents flags

```
RunFocus(["-h"]) or ["--help"] -> exit 0; stdout lists focus, --index, --dry-run; trailing \n
```

## Steps

1. CLIArgs are help flags only (no session id required).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CLIArgs = []string{"-h"}
	return nil
}
```
