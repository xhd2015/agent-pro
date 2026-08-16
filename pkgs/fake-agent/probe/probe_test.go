package probe

import (
	"strings"
	"testing"
)

func TestScan_BacktickToolCall(t *testing.T) {
	suggestions := Scan("please run `echo yes` for me")
	if !hasToolSuggestion(suggestions, "echo yes") {
		t.Fatalf("expected tool_call:echo yes, got %+v", suggestions)
	}
}

func TestScan_DollarToolCall(t *testing.T) {
	suggestions := Scan("do this:\n$ ls -la\nand check")
	if !hasToolSuggestion(suggestions, "ls -la") {
		t.Fatalf("expected tool_call:ls -la, got %+v", suggestions)
	}
}

func TestScan_RunBacktickToolCall(t *testing.T) {
	suggestions := Scan("I want you to run `go build ./...` now")
	if !hasToolSuggestion(suggestions, "go build ./...") {
		t.Fatalf("expected tool_call:go build ./..., got %+v", suggestions)
	}
}

func TestScan_AbsolutePath(t *testing.T) {
	suggestions := Scan("check the file at /tmp/log.txt for errors")
	if !hasFileReadSuggestion(suggestions, "/tmp/log.txt") {
		t.Fatalf("expected file_read:/tmp/log.txt, got %+v", suggestions)
	}
}

func TestScan_HomePath(t *testing.T) {
	suggestions := Scan("read ~/.config/settings.json")
	if !hasFileReadSuggestion(suggestions, "~/.config/settings.json") {
		t.Fatalf("expected file_read:~/.config/settings.json, got %+v", suggestions)
	}
}

func TestScan_RelativePath(t *testing.T) {
	suggestions := Scan("look at ./src/main.go")
	if !hasFileReadSuggestion(suggestions, "./src/main.go") {
		t.Fatalf("expected file_read:./src/main.go, got %+v", suggestions)
	}
}

func TestScan_ReadVerb(t *testing.T) {
	suggestions := Scan("please read README.md and check the contents")
	if !hasFileReadSuggestion(suggestions, "README.md") {
		t.Fatalf("expected file_read:README.md, got %+v", suggestions)
	}
}

func TestScan_CatVerb(t *testing.T) {
	suggestions := Scan("can you cat /var/log/syslog")
	if !hasFileReadSuggestion(suggestions, "/var/log/syslog") {
		t.Fatalf("expected file_read:/var/log/syslog, got %+v", suggestions)
	}
}

func TestScan_WriteVerb(t *testing.T) {
	suggestions := Scan("create output.txt with the results")
	if !hasFileWriteSuggestion(suggestions, "output.txt") {
		t.Fatalf("expected file_write:output.txt, got %+v", suggestions)
	}
}

func TestScan_WriteVerbPath(t *testing.T) {
	suggestions := Scan("write /tmp/config.json")
	if !hasFileWriteSuggestion(suggestions, "/tmp/config.json") {
		t.Fatalf("expected file_write:/tmp/config.json, got %+v", suggestions)
	}
}

func TestScan_SaveToVerb(t *testing.T) {
	suggestions := Scan("save to /tmp/result.json")
	if !hasFileWriteSuggestion(suggestions, "/tmp/result.json") {
		t.Fatalf("expected file_write:/tmp/result.json, got %+v", suggestions)
	}
}

func TestScan_SearchVerb(t *testing.T) {
	suggestions := Scan("search for error patterns in the logs")
	if !hasSearchSuggestion(suggestions, "error patterns in the logs") {
		t.Fatalf("expected search, got %+v", suggestions)
	}
}

func TestScan_GrepVerb(t *testing.T) {
	suggestions := Scan("grep TODO src/")
	if !hasSearchSuggestion(suggestions, "TODO src/") {
		t.Fatalf("expected search, got %+v", suggestions)
	}
}

func TestScan_MultipleMatches(t *testing.T) {
	suggestions := Scan("read /tmp/a.txt and create /tmp/b.txt and search for hello")
	found := 0
	for _, s := range suggestions {
		switch {
		case s.Kind == KindFileRead && s.Value == "/tmp/a.txt":
			found++
		case s.Kind == KindFileWrite && s.Value == "/tmp/b.txt":
			found++
		case s.Kind == KindSearch:
			found++
		}
	}
	if found < 3 {
		t.Fatalf("expected at least 3 matches, got %d: %+v", found, suggestions)
	}
}

func TestScan_NoMatch(t *testing.T) {
	suggestions := Scan("hello world")
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions, got %+v", suggestions)
	}
}

func TestScan_EmptyInput(t *testing.T) {
	suggestions := Scan("")
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions for empty input, got %+v", suggestions)
	}
}

func TestScan_Deterministic(t *testing.T) {
	input := "run `echo yes` and read /tmp/log.txt"
	result1 := Scan(input)
	result2 := Scan(input)
	if len(result1) != len(result2) {
		t.Fatalf("deterministic: len1=%d, len2=%d", len(result1), len(result2))
	}
	for i := range result1 {
		if result1[i] != result2[i] {
			t.Fatalf("deterministic: pos %d differs: %+v vs %+v", i, result1[i], result2[i])
		}
	}
}

func TestScan_Deduplicates(t *testing.T) {
	suggestions := Scan("run `echo yes` and `echo yes` again")
	count := 0
	for _, s := range suggestions {
		if s.Kind == KindToolCall && s.Value == "echo yes" {
			count++
		}
	}
	if count > 1 {
		t.Fatalf("expected deduplication, got %d occurrences of echo yes", count)
	}
}

func TestScan_FiltersCommonWords(t *testing.T) {
	suggestions := Scan("read the file and write a test")
	for _, s := range suggestions {
		if s.Value == "the" || s.Value == "a" || s.Value == "and" {
			t.Fatalf("expected common words filtered out, got %+v", suggestions)
		}
	}
}

func TestMerge_EmptyCurrent(t *testing.T) {
	incoming := []Suggestion{{KindFileRead, "a.txt"}}
	result := Merge(nil, incoming)
	if len(result) != 1 || result[0].Value != "a.txt" {
		t.Fatalf("Merge(nil, incoming) = %+v", result)
	}
}

func TestMerge_EmptyIncoming(t *testing.T) {
	current := []Suggestion{{KindFileRead, "a.txt"}}
	result := Merge(current, nil)
	if len(result) != 1 || result[0].Value != "a.txt" {
		t.Fatalf("Merge(current, nil) = %+v", result)
	}
}

func TestMerge_Deduplicates(t *testing.T) {
	current := []Suggestion{{KindFileRead, "a.txt"}, {KindToolCall, "ls"}}
	incoming := []Suggestion{{KindFileRead, "a.txt"}, {KindFileWrite, "b.txt"}}
	result := Merge(current, incoming)
	if len(result) != 3 {
		t.Fatalf("expected 3 merged, got %d: %+v", len(result), result)
	}
}

func TestMerge_PreservesOrder(t *testing.T) {
	current := []Suggestion{{KindFileRead, "a.txt"}, {KindToolCall, "ls"}}
	incoming := []Suggestion{{KindFileWrite, "b.txt"}, {KindSearch, "query"}}
	result := Merge(current, incoming)
	if len(result) != 4 {
		t.Fatalf("expected 4, got %d", len(result))
	}
	if result[0].Value != "a.txt" || result[1].Value != "ls" || result[2].Value != "b.txt" || result[3].Value != "query" {
		t.Fatalf("order not preserved: %+v", result)
	}
}

func TestDefaultSuggestions(t *testing.T) {
	defs := DefaultSuggestions()
	if len(defs) == 0 {
		t.Fatal("DefaultSuggestions returned empty")
	}
	seen := make(map[string]bool)
	for _, s := range defs {
		if s.Value == "" || s.Kind == "" {
			t.Fatalf("invalid default suggestion: %+v", s)
		}
		if seen[s.Value] {
			t.Fatalf("duplicate default suggestion: %s", s.Value)
		}
		seen[s.Value] = true
		if strings.HasPrefix(s.Value, "go ") {
			t.Fatalf("default suggestion %q would nest a go command under go test", s.Value)
		}
	}
}

func hasToolSuggestion(suggestions []Suggestion, value string) bool {
	for _, s := range suggestions {
		if s.Kind == KindToolCall && s.Value == value {
			return true
		}
	}
	return false
}

func hasFileReadSuggestion(suggestions []Suggestion, value string) bool {
	for _, s := range suggestions {
		if s.Kind == KindFileRead && s.Value == value {
			return true
		}
	}
	return false
}

func hasFileWriteSuggestion(suggestions []Suggestion, value string) bool {
	for _, s := range suggestions {
		if s.Kind == KindFileWrite && s.Value == value {
			return true
		}
	}
	return false
}

func hasSearchSuggestion(suggestions []Suggestion, value string) bool {
	for _, s := range suggestions {
		if s.Kind == KindSearch && s.Value == value {
			return true
		}
	}
	return false
}
