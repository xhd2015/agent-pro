# Scenario

**Feature**: two positional messages rejected

```
Caller -> slack-send ... "one" "two" -> exactly one message required
```

## Steps

1. Append two message positionals after flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = append(req.Args, "first", "second")
	return nil
}
```