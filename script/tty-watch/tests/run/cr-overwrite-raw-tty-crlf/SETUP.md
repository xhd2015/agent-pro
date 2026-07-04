# Scenario

**Bug**: carriage-return overwrite on raw TTY must preserve `\r` while normalizing bare LF to CRLF

```
harness PTY -> tty-watch run sh -c "printf 'MARKER_A\\rMARKER_B\\n'"
  -> MARKER_B visible, not MARKER_AMARKER_B; wire format has no bare LF newlines
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-cr-overwrite"
	req.RunCommand = []string{"sh", "-c", `printf 'MARKER_A\rMARKER_B\n'`}
	return nil
}
```