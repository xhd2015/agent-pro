# Scenario

**Feature**: EnsureInstalled errors when stdin is not a TTY

```
# not installed + !stdin TTY -> error before visudo/install
```

## Preconditions

- No prior install.
- Stdin is not an interactive terminal.

## Steps

1. Set `StdinIsTerminal = false`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "non_tty_requires_terminal"
	req.StdinIsTerminal = false
	return nil
}
```