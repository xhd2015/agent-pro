---
label: e2e
---

## Expected

- The `listen_addr` from the persisted registry entry is reachable while the `--keep-tty` process is running.
- The registry entry has valid `listen_addr` and `pid` fields.

## Side Effects

- Background `agent-run run --keep-tty` started during Setup; killed on test cleanup.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.RegistryEntry == nil {
		t.Fatal("expected registry entry while keep-tty run is active")
	}
	entry := resp.RegistryEntry
	if entry.ListenAddr == "" {
		t.Fatal("registry entry has empty listen_addr")
	}
	if !portOpen(entry.ListenAddr) {
		t.Fatalf("listen_addr %s not reachable during --keep-tty run", entry.ListenAddr)
	}
}
```
