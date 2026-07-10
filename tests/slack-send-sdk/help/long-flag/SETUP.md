# Scenario

**Feature**: `--help` prints full usage

```
slack-send --help -> usage block on stdout -> exit 0
```

## Steps

1. Set `req.Args` to `["--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```