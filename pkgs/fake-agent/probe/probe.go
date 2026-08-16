package probe

import (
	"regexp"
	"strings"
)

type Kind string

const (
	KindToolCall  Kind = "tool_call"
	KindFileRead  Kind = "file_read"
	KindFileWrite Kind = "file_write"
	KindSearch    Kind = "search"
)

type Suggestion struct {
	Kind  Kind
	Value string
}

type pattern struct {
	kind    Kind
	re      *regexp.Regexp
	trimIdx int
}

var patterns = []pattern{
	{KindToolCall, regexp.MustCompile("`([^`]+)`"), 1},
	{KindToolCall, regexp.MustCompile(`(?m)^\$\s+(.+)$`), 1},
	{KindToolCall, regexp.MustCompile(`(?:run|execute|Run|Execute)\s+` + "`" + `([^` + "`" + `]+)` + "`"), 1},

	{KindFileRead, regexp.MustCompile(`(?:\s|^)(/(?:\w[-.\w]*/)*\w[-.\w]*)`), 1},
	{KindFileRead, regexp.MustCompile(`(?:\s|^)(~/[\w./-]+)`), 1},
	{KindFileRead, regexp.MustCompile(`(?:\s|^)(\.\.?/[\w./-]+)`), 1},

	{KindFileRead, regexp.MustCompile(`(?i)\b(?:read|cat|open|less|view)\s+(\S+)`), 1},

	{KindFileWrite, regexp.MustCompile(`(?i)\b(?:create|write|touch|make|edit|modify)\s+(\S+)`), 1},
	{KindFileWrite, regexp.MustCompile(`(?i)(?:save|output)\s+to\s+(\S+)`), 1},

	{KindSearch, regexp.MustCompile(`(?i)\b(?:search\s+for|find)\s+(\S.+)`), 1},
	{KindSearch, regexp.MustCompile(`(?i)\bgrep\s+(\S.+)`), 1},
}

var defaultSuggestions = []Suggestion{
	{KindToolCall, "ls -la"},
	{KindToolCall, "git status"},
	// Cheap, non-go probe. Do not default to `go test` / `go env`:
	// GenerateSession actually execs these, and a nested `go` waits on
	// the same command lock as the parent `go test`.
	{KindToolCall, "pwd"},
	{KindFileRead, "README.md"},
	{KindSearch, "TODO"},
}

func DefaultSuggestions() []Suggestion {
	result := make([]Suggestion, len(defaultSuggestions))
	copy(result, defaultSuggestions)
	return result
}

func Scan(text string) []Suggestion {
	var result []Suggestion
	seen := make(map[string]bool)

	for _, p := range patterns {
		matches := p.re.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			if len(m) <= p.trimIdx {
				continue
			}
			value := strings.TrimSpace(m[p.trimIdx])
			if value == "" {
				continue
			}
			if isCommonWord(value) {
				continue
			}
			key := string(p.kind) + ":" + normalizeValue(value)
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, Suggestion{Kind: p.kind, Value: value})
		}
	}

	return result
}

func Merge(current, incoming []Suggestion) []Suggestion {
	if len(incoming) == 0 {
		return current
	}
	if len(current) == 0 {
		result := make([]Suggestion, len(incoming))
		copy(result, incoming)
		return result
	}

	seen := make(map[string]bool)
	for _, s := range current {
		seen[suggestionKey(s)] = true
	}

	result := make([]Suggestion, len(current))
	copy(result, current)

	for _, s := range incoming {
		key := suggestionKey(s)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, s)
	}

	return result
}

func suggestionKey(s Suggestion) string {
	return string(s.Kind) + ":" + normalizeValue(s.Value)
}

func normalizeValue(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

var commonWords = map[string]bool{
	"a": true, "an": true, "the": true,
	"is": true, "are": true, "was": true, "were": true,
	"be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true,
	"will": true, "would": true, "shall": true, "should": true,
	"may": true, "might": true, "must": true, "can": true, "could": true,
	"in": true, "on": true, "at": true, "to": true, "for": true,
	"of": true, "from": true, "with": true, "by": true,
	"this": true, "that": true, "these": true, "those": true,
	"it": true, "its": true, "i": true, "you": true, "he": true, "she": true,
	"we": true, "they": true, "me": true, "him": true, "her": true, "us": true, "them": true,
	"and": true, "or": true, "not": true, "but": true, "if": true, "so": true,
	"then": true, "here": true, "there": true, "all": true, "some": true, "any": true,
	"no": true, "up": true, "down": true, "out": true, "about": true,
	"just": true, "now": true, "very": true, "also": true, "only": true,
	"please": true, "need": true, "want": true, "like": true,
	"file": true, "files": true, "code": true, "test": true, "tests": true,
	"fix": true, "add": true, "new": true, "use": true, "using": true,
	"change": true, "changes": true, "make": true, "let": true, "get": true,
}

func isCommonWord(s string) bool {
	return commonWords[strings.ToLower(strings.TrimSpace(s))]
}
