# Scenario

**Feature**: extract user text from grok `<user_query>` wrapper for random-fallback topic

Reproduces tty-watch `Hello\r` into grok TUI: LLM request uses wrapped prompt
`<user_query>\nHello\n</user_query>` but generated events say `<user_query>` not Hello.

```
extractTopic stops at first newline -> topic "<user_query>" instead of "Hello"
```

## Steps

1. Empty `exchanges[]`.
2. One request with grok-style wrapped user content.
3. Assert think/message text references Hello, not the XML tag.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": []
}`
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"<user_query>\nHello\n</user_query>"}]}`,
	}
	return nil
}
```