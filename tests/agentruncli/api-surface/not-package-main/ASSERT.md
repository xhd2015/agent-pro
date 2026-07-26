## Expected

- At least one non-test `.go` under `pkgs/agentruncli` was scanned.
- Package name is `agentruncli` (not `main`).

## Side Effects

- None (read-only source inspection).

## Errors

- Harness error if package directory is missing (RED until dir exists).

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	// Missing pkgs/agentruncli is the Classic RED state until implementer creates it.
	if err != nil {
		t.Fatalf("pkgs/agentruncli must exist as a library package: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.ScannedFiles == 0 {
		t.Fatal("expected to scan pkgs/agentruncli .go files")
	}
	if resp.PackageName == "main" {
		t.Fatal("pkgs/agentruncli must not be package main (library extract)")
	}
	assertEqual(t, "PackageName", resp.PackageName, "agentruncli")
}
```
