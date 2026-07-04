# Scenario

**Feature**: preset `simple` (one message) drained; second HTTP serve uses genStream think fallback

```
genQueue [message] -> POST #1 -> preset message
genQueue empty -> POST #2 -> genStream ActionThink -> HTTP 200 (not no_match)
```

## Steps

1. Empty config (`exchanges: []`).
2. `--mock-events-preset=simple`.
3. Send two chat completion requests in order.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigJSON = `{"port": 8080, "exchanges": []}`
	req.MockEventsPreset = "simple"
	req.Requests = []string{
		`{"model":"mock-model","messages":[{"role":"user","content":"preset-first-prompt"}]}`,
		`{"model":"mock-model","messages":[{"role":"user","content":"genstream-second-prompt"}]}`,
	}
	return nil
}
```