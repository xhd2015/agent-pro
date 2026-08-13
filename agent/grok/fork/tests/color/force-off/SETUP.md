# Scenario

**Feature**: `--no-color` strips ANSI from Mode A success

```
fork.Main(["--no-color"]) -> Opened line, no CSI
```

## Steps

1. Args `["--no-color"]`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--no-color"}
	return nil
}
```
