# Scenario

**Feature**: `-h` prints the same usage contract as `--help`

```
fork.Main([]string{"-h"}) -> stdout usage, err=nil
```

## Steps

1. Args `["-h"]`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"-h"}
	return nil
}
```
