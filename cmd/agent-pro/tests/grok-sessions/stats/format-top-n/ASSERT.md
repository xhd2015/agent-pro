## Expected

- With `TopN=2`, output includes section header `Top tools by total handler time`.
- Numbered top-tool rows (`1` …) are at most 2 (no third ranked tool row for `read_file` as #3).
- Highest total (`search_replace`, 2×100ms) appears in the top tools section.
- Full tool table still present under `Tool handler time` (all tools).
- Top background tasks header: `#  DURATION  EXIT  COMMAND` (not LABEL-only).
- Top subagents header includes STATUS, TYPE, TOOLS, TURNS, DESC.
- Command fragment `doctest test` and spawn desc `explore stats tree` appear
  (not UUID-only for the subagent when description is present).
- Calling `FormatStatsTextOpts` again with `TopN=0` yields **no** headers:
  - `Top tools by total handler time`
  - `Top background tasks`
  - `Top subagents`

## Errors

- None.

```go
import (
	"regexp"
	"strings"
	"testing"

	sessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	out := resp.Output

	assertContains(t, out, "Tool handler time")
	assertContains(t, out, "Top tools by total handler time")
	assertContains(t, out, "search_replace")
	assertContains(t, out, "Top background tasks")
	assertContains(t, out, "Top subagents")

	// Rich bg/sub content
	assertContains(t, out, "doctest test")
	assertContains(t, out, "explore stats tree")

	// Rich headers (locked)
	assertContains(t, out, "#  DURATION  EXIT  COMMAND")
	bg := sectionAfter(out, "Top background tasks")
	if bg == "" {
		t.Fatalf("missing Top background tasks:\n%s", out)
	}
	if strings.Contains(bg, "LABEL") && !strings.Contains(bg, "COMMAND") {
		t.Fatalf("top-bg still LABEL-only:\n%s", bg)
	}

	sub := sectionAfter(out, "Top subagents")
	if sub == "" {
		t.Fatalf("missing Top subagents:\n%s", out)
	}
	// Not UUID-only when description present.
	assertContains(t, sub, "explore stats tree")
	// Header columns for rich subagent table.
	subHdrOK := strings.Contains(sub, "STATUS") && strings.Contains(sub, "TYPE") &&
		strings.Contains(sub, "TOOLS") && strings.Contains(sub, "TURNS") &&
		strings.Contains(sub, "DESC")
	if !subHdrOK {
		t.Fatalf("top-subagents header missing STATUS/TYPE/TOOLS/TURNS/DESC:\n%s", sub)
	}

	// Count numbered lines under Top tools section.
	topSection := out
	if i := strings.Index(out, "Top tools by total handler time"); i >= 0 {
		topSection = out[i:]
		for _, stop := range []string{"\nTop background", "\nTop subagents", "\nBackground tasks", "\nSubagents", "\nSources"} {
			if j := strings.Index(topSection[1:], stop); j >= 0 {
				topSection = topSection[:j+1]
				break
			}
		}
	}
	numRe := regexp.MustCompile(`(?m)^\s*1\s+`)
	num2Re := regexp.MustCompile(`(?m)^\s*2\s+`)
	num3Re := regexp.MustCompile(`(?m)^\s*3\s+`)
	if !numRe.MatchString(topSection) {
		t.Fatalf("Top tools missing numbered row 1:\n%s", topSection)
	}
	if !num2Re.MatchString(topSection) {
		t.Fatalf("Top tools missing numbered row 2 (TopN=2):\n%s", topSection)
	}
	if num3Re.MatchString(topSection) {
		t.Fatalf("Top tools has row 3 but TopN=2:\n%s", topSection)
	}
	if idxSR := strings.Index(topSection, "search_replace"); idxSR < 0 {
		t.Fatalf("Top tools missing search_replace:\n%s", topSection)
	}

	// TopN=0 hides all Top section headers.
	hidden := sessions.FormatStatsTextOpts(resp.Stats, sessions.FormatStatsOptions{
		Home:      req.GrokHome,
		Now:       req.Now,
		ColorMode: "never",
		TopN:      0,
	})
	for _, hdr := range []string{
		"Top tools by total handler time",
		"Top background tasks",
		"Top subagents",
	} {
		if strings.Contains(hidden, hdr) {
			t.Fatalf("TopN=0 should hide %q, got:\n%s", hdr, hidden)
		}
	}
}
```
