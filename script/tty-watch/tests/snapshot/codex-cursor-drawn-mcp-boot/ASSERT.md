## Expected

- Exit code 0.
- Snapshot contains the **full** deprecation warning starting with `codex_hooks deprecated`.
- Snapshot contains the tip line `Tip: New Use /fast` once.
- Snapshot shows `OpenAI Codex` once (latest screen only).
- Snapshot shows final MCP error once; no stacked `Starting MCP servers` boot smear.
- Snapshot shows `Write tests for @filename` once (current prompt).
- Status bar `gpt-5.5 medium ·` appears once.
- No kitty/CSI leak fragments like `[<` or `;23M`.
- No smeared long lines (max line length 120).
- No CSI/OSC/C0 escape leaks in output.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("snapshot exit %d stdout %q stderr %q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	text := resp.SnapshotText
	if text == "" {
		text = resp.Stdout
	}
	if !strings.Contains(text, "codex_hooks deprecated") {
		t.Fatalf("snapshot truncated top warning (want full deprecation line), got %q", text)
	}
	if strings.Count(text, "Tip: New Use /fast") != 1 {
		t.Fatalf("snapshot tip missing or stacked (want 1), got %q", text)
	}
	if count := strings.Count(text, "OpenAI Codex"); count != 1 {
		t.Fatalf("snapshot stacked %d UI headers (want 1 latest screen), got %q", count, text)
	}
	if strings.Count(text, "Starting MCP servers") > 0 {
		t.Fatalf("snapshot smeared MCP boot progress (want final error only), got %q", text)
	}
	if strings.Count(text, "MCP startup incomplete") != 1 {
		t.Fatalf("snapshot missing final MCP error once, got %q", text)
	}
	if count := strings.Count(text, "Write tests for @filename"); count != 1 {
		t.Fatalf("snapshot stacked %d prompt lines (want 1), got %q", count, text)
	}
	if count := strings.Count(text, "gpt-5.5 medium ·"); count != 1 {
		t.Fatalf("snapshot stacked %d status bars (want 1), got %q", count, text)
	}
	if strings.Contains(text, "[<") || strings.Contains(text, ";23M") {
		t.Fatalf("snapshot leaked kitty/CSI fragments, got %q", text)
	}
	if resp.ContainsEscape {
		t.Fatalf("snapshot still contains escape sequences: %q", text)
	}
	for _, line := range strings.Split(text, "\n") {
		if len(line) > 120 {
			t.Fatalf("snapshot smeared long line (%d chars) from partial redraw overlap: %q", len(line), line)
		}
	}
}
```