package sessions

import (
	"encoding/json"
	"strings"
	"testing"
)

// bashCall builds the tool/call event shape DSH writes: `arguments` holds a
// JSON string rather than a JSON object.
func bashCall(seq int64, args map[string]any) Event {
	encoded, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}
	quoted, err := json.Marshal(string(encoded))
	if err != nil {
		panic(err)
	}
	return Event{Type: "tool/call", Seq: seq, Time: seq, Data: json.RawMessage(`{"name":"bash","arguments":` + string(quoted) + `}`)}
}

// The transcript is where a row's Text comes from; a double-encoded arguments
// payload used to fall through the map unmarshal and leave Text empty.
func TestToolCallPreviewDecodesStringArguments(t *testing.T) {
	event := bashCall(1, map[string]any{
		"command":             "pwd; /usr/bin/mdfind 'kMDItemFSName == \"*Marcus*\"cd'",
		"description":         "Locate installed Marcus app",
		"sandbox_permissions": "danger-full-access",
		"workdir":             "/tmp",
	})
	name, text := toolCallPreview(event)
	if name != "bash" {
		t.Fatalf("name=%q", name)
	}
	if want := "pwd; /usr/bin/mdfind 'kMDItemFSName == \"*Marcus*\"cd'"; text != want {
		t.Fatalf("text=%q want %q", text, want)
	}
}

// An object payload must keep decoding after the shared decoder took over.
func TestToolCallPreviewKeepsObjectArguments(t *testing.T) {
	event := Event{Type: "tool/call", Seq: 1, Time: 1, Data: json.RawMessage(`{"name":"bash","arguments":{"command":"ls -la"}}`)}
	if _, text := toolCallPreview(event); text != "ls -la" {
		t.Fatalf("text=%q", text)
	}
}

func bashRow(t *testing.T, event Event) Message {
	t.Helper()
	rows := Project([]Event{event})
	if len(rows) != 1 || rows[0].Tool != "bash" {
		t.Fatalf("rows=%+v", rows)
	}
	return rows[0]
}

// The reported bug: a bash row summarized as its sandbox policy instead of the
// command it ran. Text carries the command, Sum must agree with it.
func TestToolSummaryBashPrefersCommand(t *testing.T) {
	row := bashRow(t, bashCall(325, map[string]any{
		"command":             "pwd; /usr/bin/mdfind 'kMDItemContentType == \"com.apple.application-bundle\"'\nopen -a Marcus",
		"description":         "Locate installed Marcus app and check running processes",
		"justification":       "",
		"run_in_background":   false,
		"sandbox_permissions": "danger-full-access",
		"timeoutMs":           10000,
		"workdir":             "/Users/xhd2015/Projects/xhd2015/working/ai",
	}))
	if !strings.HasPrefix(row.Text, "pwd; /usr/bin/mdfind") {
		t.Fatalf("text=%q", row.Text)
	}
	if want := "pwd; /usr/bin/mdfind 'kMDItemContentType == \"com.apple.application-bundle\"'"; row.Sum != want {
		t.Fatalf("sum=%q want %q", row.Sum, want)
	}
}

// The summary is a pure function of the event: reading one row repeatedly must
// never walk into a different argument.
func TestToolSummaryIsDeterministic(t *testing.T) {
	event := bashCall(24, map[string]any{
		"command":             "pwd; browser-agent session new",
		"sandbox_permissions": "danger-full-access",
		"workdir":             "/Users/xhd2015/Projects/xhd2015/working/ai",
	})
	first := bashRow(t, event).Sum
	for i := 0; i < 64; i++ {
		if got := bashRow(t, event).Sum; got != first {
			t.Fatalf("iteration %d: sum=%q want %q", i, got, first)
		}
	}
	if first != "pwd; browser-agent session new" {
		t.Fatalf("sum=%q", first)
	}
}

// A call that carries only bookkeeping arguments has nothing to show; guessing
// one of them is what produced `danger-full-access` rows.
func TestToolSummaryIgnoresBookkeepingArguments(t *testing.T) {
	row := bashRow(t, bashCall(9, map[string]any{
		"justification":       "the user asked for it",
		"sandbox_permissions": "danger-full-access",
		"timeoutMs":           10000,
		"workdir":             "/Users/xhd2015/Projects/xhd2015/working/ai",
	}))
	if row.Sum != "" || row.Text != "" {
		t.Fatalf("sum=%q text=%q", row.Sum, row.Text)
	}
}

// Tools without a dedicated case fall back to a fixed key order, and the
// policy fields stay out of it.
func TestToolSummaryFallbackSkipsPolicyFields(t *testing.T) {
	event := Event{Type: "tool/call", Seq: 18, Time: 18, Data: json.RawMessage(
		`{"name":"skill","arguments":{"name":"browser-agent","sandbox_permissions":"danger-full-access","workdir":"/tmp"}}`)}
	rows := Project([]Event{event})
	if len(rows) != 1 || rows[0].Sum != "browser-agent" {
		t.Fatalf("rows=%+v", rows)
	}
	unknown := Event{Type: "tool/call", Seq: 19, Time: 19, Data: json.RawMessage(
		`{"name":"mystery_tool","arguments":{"note":"hello world","workdir":"/tmp"}}`)}
	rows = Project([]Event{unknown})
	if len(rows) != 1 || rows[0].Sum != "hello world" {
		t.Fatalf("unknown tool rows=%+v", rows)
	}
}

// Descriptions still surface when a shell call carries no command text.
func TestToolSummaryBashFallsBackToDescription(t *testing.T) {
	row := bashRow(t, bashCall(2, map[string]any{
		"description":         "Restart the dev server",
		"sandbox_permissions": "danger-full-access",
	}))
	if row.Sum != "Restart the dev server" {
		t.Fatalf("sum=%q", row.Sum)
	}
}
