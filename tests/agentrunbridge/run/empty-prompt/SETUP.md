# Scenario

**Feature**: empty / whitespace-only prompt rejects before exec

```
Run(Prompt="" | "   ") -> error; LookPathCalls==0; LaunchCalls==0
```

## Preconditions

- Prompt required non-empty after trim.
- Session and binary otherwise valid.

## Steps

1. Set Prompt to whitespace-only.
2. Leave hooks as default success (must not be called).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "   \t  "
	req.SessionID = "sess-empty-prompt"
	req.KeepTTY = true
	return nil
}
```
