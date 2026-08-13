# Scenario

**Feature**: `--help` prints usage and exits 0

```
fork.Main([]string{"--help"}) -> stdout usage, err=nil
```

## Steps

1. Args `["--help"]`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--help"}
	return nil
}
```
