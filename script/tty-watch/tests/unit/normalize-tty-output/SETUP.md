# Scenario

**Bug**: normalizeTTYOutput must expand bare LF to CRLF while preserving standalone CR

```
LF-only screen snapshot text -> normalizeTTYOutput -> CRLF on raw TTY stdout
standalone \r (in-place redraw) -> unchanged; trailing bare \n -> \r\n
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "unit-normalize-tty-output"
	return nil
}
```