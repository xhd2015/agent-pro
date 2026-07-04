# Scenario

**Feature**: snapshot prints sanitized text without escape sequences

```
# session prints ANSI red + plain line
harness -> detached printf ANSI -> snapshot -> PLAIN_LINE without escapes
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "snapshot-sanitize"
	return nil
}
```