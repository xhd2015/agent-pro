## Expected

- No API error.
- Launch argv includes:
  - `-e` + `SLACK_MSG_SESSION_ID=sess-io-env`
  - `-e` + `SLACK_MSG_CONFIG=/abs/config.json`
- Open profile flags still present (`--open`, `--auto-send-or-resume`).
- Env tokens appear before `--` separator when open profile is used.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertHasArg(t, resp.Args, "--open")
	assertHasArg(t, resp.Args, "--auto-send-or-resume")
	// Locate -e pairs
	idxSID := assertArgIndex(t, resp.Args, "SLACK_MSG_SESSION_ID=sess-io-env")
	if idxSID == 0 || resp.Args[idxSID-1] != "-e" {
		t.Fatalf("expected -e before SLACK_MSG_SESSION_ID, argv=%q", resp.Args)
	}
	idxCfg := assertArgIndex(t, resp.Args, "SLACK_MSG_CONFIG=/abs/config.json")
	if idxCfg == 0 || resp.Args[idxCfg-1] != "-e" {
		t.Fatalf("expected -e before SLACK_MSG_CONFIG, argv=%q", resp.Args)
	}
	dashDash := assertArgIndex(t, resp.Args, "--")
	if idxSID >= dashDash || idxCfg >= dashDash {
		t.Fatalf("env flags must appear before --, argv=%q", resp.Args)
	}
}
```
