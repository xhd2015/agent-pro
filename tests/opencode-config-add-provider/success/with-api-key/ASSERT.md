## Expected

- Exit code 0.
- `provider.keyprov.options.apiKey` == `mysecret` (written inline in plaintext).
- `provider.keyprov.options.baseURL` == `https://api.example.com/v1` (still correct alongside `apiKey`).
- `provider.keyprov.npm` == `@ai-sdk/anthropic`.
- `provider.keyprov.name` == `keyprov` (defaults to id).
- `provider.keyprov.models.m1.name` == `m1`.

## Side Effects

- Global config file created with the provider entry whose `options` map contains both `baseURL` and `apiKey`.

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
	// Shared success invariants (exit 0, id mentioned, provider entry with
	// npm/options/models present). Inherited from success/SETUP.md.
	assertSuccessCommon(t, req, resp, "keyprov")

	entry, _, e := readProviderEntry(resp.ConfigPath, "keyprov")
	if e != nil {
		t.Fatal(e)
	}
	if got := entry["npm"]; got != "@ai-sdk/anthropic" {
		t.Fatalf("npm = %v, want @ai-sdk/anthropic", got)
	}
	if got := entry["name"]; got != "keyprov" {
		t.Fatalf("name = %v, want keyprov", got)
	}
	opts, _ := entry["options"].(map[string]interface{})
	if opts == nil {
		t.Fatalf("provider.keyprov.options missing or not a map: %v", entry["options"])
	}
	if got := opts["baseURL"]; got != "https://api.example.com/v1" {
		t.Fatalf("options.baseURL = %v, want https://api.example.com/v1", got)
	}
	if got := opts["apiKey"]; got != "mysecret" {
		t.Fatalf("options.apiKey = %v, want mysecret", got)
	}
	models, _ := entry["models"].(map[string]interface{})
	m, ok := models["m1"]
	if !ok {
		t.Fatalf("models.m1 missing: %v", models)
	}
	mm, _ := m.(map[string]interface{})
	if mm["name"] != "m1" {
		t.Fatalf("models.m1.name = %v, want m1", mm["name"])
	}
}
```
