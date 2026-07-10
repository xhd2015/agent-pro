# Scenario

**Feature**: group ID with G prefix passes through

```
resolveChannel("G12345ABCDE") -> G12345ABCDE
```

## Steps

1. Args `["G12345ABCDE"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"G12345ABCDE"}
	return nil
}
```