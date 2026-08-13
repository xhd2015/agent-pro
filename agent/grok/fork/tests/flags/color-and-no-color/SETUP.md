# Scenario

**Feature**: `--color` and `--no-color` cannot be combined

```
fork.Main(["--color", "--no-color"]) -> error
```

## Steps

1. Args both color flags.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--color", "--no-color"}
	return nil
}
```
