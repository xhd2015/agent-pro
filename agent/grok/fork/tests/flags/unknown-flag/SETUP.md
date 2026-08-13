# Scenario

**Feature**: unknown flag errors and mentions `--help`

```
fork.Main(["--not-a-real-flag"]) -> error contains --help
```

## Steps

1. Args `["--not-a-real-flag"]`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--not-a-real-flag"}
	return nil
}
```
