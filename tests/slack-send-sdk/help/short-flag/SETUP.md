# Scenario

**Feature**: `-h` prints full usage

```
slack-send -h -> usage block on stdout -> exit 0
```

## Steps

1. Set `req.Args` to `["-h"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```