## Expected

- Exit code 0.
- Snapshot shows the **latest** grok-like changelog screen (menu, bordered box, prompt, footer).
- Snapshot contains `ctrl+q` menu hint (not bare `Quit q` without `ctrl`).
- Snapshot contains box top border `╭` or normalized `-` border around changelog content.
- Snapshot contains prompt marker `❯` or `›`.
- Snapshot contains footer substring `Logged in with API key` or `Grok Build`.
- Snapshot must NOT show the buggy changelog-only layout (`Quit q` without `ctrl+q`, no box).
- Snapshot must NOT contain a standalone smear line `Quit q` (scrollback ghost from boot redraw).
- `Grok Build Changelog` header appears once (no stacked redraw smear).
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

	hasCtrlQ := strings.Contains(text, "ctrl+q")
	hasBox := strings.Contains(text, "╭") ||
		(strings.Contains(text, "Grok Build Changelog") && strings.Contains(text, "│"))
	hasPrompt := strings.Contains(text, "❯") || strings.Contains(text, "›")
	hasFooter := strings.Contains(text, "Logged in with API key") || strings.Contains(text, "Grok Build")

	if !hasCtrlQ {
		t.Fatalf("snapshot missing ctrl+q menu hint (scrollback path bug), got %q", text)
	}
	if strings.Contains(text, "Quit q") && !hasCtrlQ {
		t.Fatalf("snapshot shows garbled Quit q without ctrl+q, got %q", text)
	}
	if !hasBox {
		t.Fatalf("snapshot missing bordered changelog box (╭ or │), got %q", text)
	}
	if !hasPrompt {
		t.Fatalf("snapshot missing prompt marker ❯ or ›, got %q", text)
	}
	if !hasFooter {
		t.Fatalf("snapshot missing footer (Logged in with API key / Grok Build), got %q", text)
	}
	if strings.Count(text, "Grok Build Changelog") != 1 {
		t.Fatalf("snapshot stacked %d changelog headers (want 1 latest screen), got %q",
			strings.Count(text, "Grok Build Changelog"), text)
	}
	if strings.Contains(text, "Quit q") && !strings.Contains(text, "╭") && !strings.Contains(text, "│") {
		t.Fatalf("snapshot shows changelog-only layout without box (current bug), got %q", text)
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "Quit q" {
			t.Fatalf("snapshot shows scrollback smear line Quit q (want frame path only), got %q", text)
		}
	}
	if resp.ContainsEscape {
		t.Fatalf("snapshot still contains escape sequences: %q", text)
	}
	for _, line := range strings.Split(text, "\n") {
		if len(line) > 120 {
			t.Fatalf("snapshot smeared long line (%d chars) from stacked redraw overlap: %q", len(line), line)
		}
	}
}
```