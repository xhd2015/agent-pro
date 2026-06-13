## Preconditions
- An event with an invalid event_type is sent.

## Steps
1. Run `agent-hub notify --json '{"event_type":"bogus","runner":"x"}'`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"notify", "--json", `{"event_type":"bogus","runner":"x"}`}
    return nil
}
```
