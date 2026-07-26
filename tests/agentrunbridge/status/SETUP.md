# Scenario

**Feature**: pure TTY status parsing helpers

```
agent-run tty status stdout
  -> ParseTTYStatus -> (screen, sendable)
  -> IsSessionReady -> banner+yes?
```

## Preconditions

- Human status lines use prefixes `screen status:` and `sendable:`.
- Comparison is case-insensitive for ready check.
- `sendable` takes first whitespace-separated token only.

## Steps

1. Set `req.Mode = "status"`.
2. Leaf sets `req.StatusStdout` fixture text.
3. Assert on `Response.Screen`, `Sendable`, `Ready`.

## Context

- Shared fixtures: `statusReadyFixture` / `statusNotReadyFixture` in root DOCTEST.md.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "status"
	return nil
}
```
