## Expected

- Startup stderr contains `agent-run web token:` followed by a hex token.
- `GET /api/agent-run/health` without `Authorization` → **401**.
- Same request with `Authorization: Bearer <printed token>` → **200**.
- `auth.token` under `AGENT_RUN_HOME` matches the printed token.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assert.Output(t, req.WebServerStderr, `<contains>
agent-run web token:
</contains>`)

	token := parseWebTokenFromStderr(req.WebServerStderr)
	if token == "" {
		t.Fatalf("could not parse token from stderr:\n%s", req.WebServerStderr)
	}
	if resp.HTTPStatus != 401 {
		t.Fatalf("expected HTTP 401 without bearer, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}

	url := req.WebBaseURL + "/api/agent-run/health"
	status, _ := httpGet(t, url, token)
	if status != 200 {
		t.Fatalf("expected HTTP 200 with auto token, got %d", status)
	}

	authPath := filepath.Join(req.Home, "auth.token")
	data, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("read auth.token: %v", err)
	}
	if strings.TrimSpace(string(data)) != token {
		t.Fatalf("auth.token = %q, want %q", strings.TrimSpace(string(data)), token)
	}
}
```