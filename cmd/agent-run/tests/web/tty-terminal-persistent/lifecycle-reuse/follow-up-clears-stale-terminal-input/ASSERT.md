## Expected

- The follow-up POST is accepted.
- The backend terminal input line is cleared/replaced before the follow-up is
  submitted.
- The chat transcript must not contain `MALFORMED_SUBMISSION`.
- The chat transcript must not contain the stale input `Explain this codebase`
  as part of the follow-up assistant response.

## Exit Code

- Test process exits non-zero while the follow-up is appended to stale terminal
  input and submitted as one combined prompt.

```go
import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.FollowUpStatus != http.StatusAccepted {
		t.Fatalf("follow-up status=%d body=%s", resp.FollowUpStatus, resp.FollowUpBody)
	}
	body := staleInputSessionBodyEventually(t, req, 5_000_000_000)
	if bad := staleMalformedAssistant(t, body); bad != "" {
		t.Fatalf("follow-up must clear stale terminal input before submit; malformed assistant=%q\nbody=%s", bad, body)
	}
	if !strings.Contains(body, "FOLLOWUP_RESPONSE") {
		input := ""
		if req.PTYInputSeen != nil {
			input = *req.PTYInputSeen
		}
		if strings.Contains(input, req.FollowUpPrompt) && !strings.Contains(input, "\x15") {
			t.Fatalf("follow-up was written without clearing stale terminal input first; pty input=%q body=%s", input, body)
		}
		t.Fatalf("follow-up should be submitted as a clean prompt and produce FOLLOWUP_RESPONSE; pty input=%q body=%s", input, body)
	}
}

func staleMalformedAssistant(t *testing.T, body string) string {
	t.Helper()
	var parsed struct {
		Events []struct {
			Type string `json:"type"`
			Role string `json:"role"`
			Text string `json:"text"`
		} `json:"events"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("invalid session detail JSON: %v\n%s", err, body)
	}
	for _, ev := range parsed.Events {
		if ev.Type != "message" || ev.Role != "assistant" {
			continue
		}
		if strings.Contains(ev.Text, "MALFORMED_SUBMISSION") ||
			strings.Contains(ev.Text, "Explain this codebase") ||
			strings.Contains(ev.Text, "Explain this codebasewhat did I say?") {
			return ev.Text
		}
	}
	return ""
}
```
