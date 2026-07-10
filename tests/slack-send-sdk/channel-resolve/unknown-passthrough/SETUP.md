# Scenario

**Feature**: unknown channel name passes through unchanged

```
resolveChannel("#unknown-channel") -> "#unknown-channel" (no map hit)
```

## Steps

1. Args `["#unknown-channel"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"#unknown-channel"}
	return nil
}
```