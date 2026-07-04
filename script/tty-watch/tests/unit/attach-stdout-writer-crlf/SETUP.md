# Scenario

**Feature**: attachStdoutWriter applies normalizeTTYOutput on interactive TTY stdout

```
screen snapshot text (LF-only) -> attachStdoutWriter rawTTY path -> CRLF on stdout
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "unit-attach-stdout-writer-crlf"
	return nil
}
```