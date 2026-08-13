# Scenario

**Feature**: `--color --dry-run` paints gray labels on the Mode A plan

```
fork.Main(["--color", "--dry-run"]) -> gray "ancestor pid:" / "grok id:" / …
```

## Steps

1. Args `["--color", "--dry-run"]`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--color", "--dry-run"}
	return nil
}
```
