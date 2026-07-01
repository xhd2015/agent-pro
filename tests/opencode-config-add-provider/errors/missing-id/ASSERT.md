## Expected

- Non-zero exit code.
- stderr mentions `--id`.

## Side Effects

- No config file written at the global target.

## Errors

- A validation error naming the missing `--id` flag.

## Exit Code

- Non-zero.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertError(t, resp, "--id")
	assertNoConfigFile(t, resp)
}
```
