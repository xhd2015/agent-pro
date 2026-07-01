## Expected

- Exit code 0.
- `provider.prov-a.name` == `My Display` (the `--name` value, not the id).
- `provider.prov-a.npm` == `@ai-sdk/anthropic`.

## Side Effects

- Global config file created with the named provider entry.

## Errors

- None.

## Exit Code

- 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessCommon(t, req, resp, "prov-a")
	entry, _, e := readProviderEntry(resp.ConfigPath, "prov-a")
	if e != nil {
		t.Fatal(e)
	}
	if got := entry["name"]; got != "My Display" {
		t.Fatalf("name = %v, want \"My Display\"", got)
	}
}
```
