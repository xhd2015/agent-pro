---
label: e2e
---

## Expected Output

```text
<contains>
no API token
--token
</contains>
```

## Expected

- Startup stderr mentions that no API token is configured and suggests `--token` (or `--token auto`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assert.Output(t, req.WebServerStderr, `<contains>
no API token
--token
</contains>`)
}
```