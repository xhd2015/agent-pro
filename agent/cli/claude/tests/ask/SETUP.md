# Scenario

**Feature**: ClaudeAgent Ask() operation routes a headless claude query and parses its stream-json

```
# Ask() builds claude args and streams stdout line by line
ask -> ClaudeAgent.Ask -> claude -p <q> --output-format stream-json --verbose
# assistant text blocks accumulate into the answer; session_id captured
ClaudeAgent <- claude (assistant text, system init session_id, result session_id)
```

## Preconditions
- The `claude` binary is available in PATH.
- This subtree covers the `Ask()` operation mode of ClaudeAgent.

## Steps
1. Set `Operation` to `"ask"` to route through the Ask path.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = OpAsk
	return nil
}
```
