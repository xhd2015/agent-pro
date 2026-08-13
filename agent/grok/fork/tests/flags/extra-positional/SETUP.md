# Scenario

**Feature**: extra positional arguments are rejected

```
fork.Main(["leftover"]) -> error
```

## Steps

1. Args `["leftover"]` (Mode A plus unexpected positional).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"leftover"}
	return nil
}
```
