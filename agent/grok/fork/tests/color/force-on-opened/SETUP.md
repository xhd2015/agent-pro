# Scenario

**Feature**: `--color` paints green `Opened` on Mode A success

```
fork.Main(["--color"]) -> stdout contains green Opened
```

## Steps

1. Args `["--color"]`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--color"}
	return nil
}
```
