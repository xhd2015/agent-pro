# Scenario

**Feature**: channel ID with C prefix passes through

```
resolveChannel("C0ALE44K5J6") -> C0ALE44K5J6
```

## Steps

1. Args `["C0ALE44K5J6", "custom text"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"C0ALE44K5J6", "custom text"}
	return nil
}
```