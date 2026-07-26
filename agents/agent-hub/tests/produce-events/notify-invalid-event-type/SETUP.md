## Preconditions
- An event with an invalid event_type is sent.

## Steps
1. Run `agent-hub notify --json '{"event_type":"bogus","runner":"x"}'`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"notify", "--json", `{"event_type":"bogus","runner":"x"}`}
    return nil
}
```
