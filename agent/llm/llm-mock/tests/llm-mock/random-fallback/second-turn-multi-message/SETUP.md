# Scenario

**Feature**: empty prefix; first turn consumes generator stream; second user turn must still random-fallback → 200

```
config exchanges[] empty
POST #1 user:Hello -> think event -> HTTP 200
POST #2 user:Hello -> message event -> HTTP 200 (stream exhausted)
POST #3 multi-turn history + user:"what's wrong with me?" -> must HTTP 200 (not no_match)
```

Reproduces grok session: first prompt "Hello" succeeds, second prompt "what's wrong with me?"
errors with `no_match` when prior turns exhausted the shared `genStream`.

## Steps

1. Config JSON with `exchanges: []`.
2. Send three chat completion requests in order (simulates grok turn 1 + turn 2).
3. Third request includes assistant reply from turn 1 and new user message.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": []
}`
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"Hello"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"Hello"}]}`,
		`{"model":"mock-model","messages":[
			{"role":"user","content":"Hello"},
			{"role":"assistant","content":"Here's the result for your request about Hello. I've made the necessary changes."},
			{"role":"user","content":"what's wrong with me?"}
		]}`,
	}
	return nil
}
```