package occupied

import "testing"

func TestExactlyOneMoreSpace(t *testing.T) {
	cases := []struct {
		name          string
		before, after string
		want          bool
	}{
		{"append trailing space", "hello", "hello ", true},
		{"insert middle space", "ab", "a b", true},
		{"prefix space", "x", " x", true},
		{"empty to one space", "", " ", false},
		{"same strings", "hello", "hello", false},
		{"two spaces added", "hi", "hi  ", false},
		{"placeholder collapse", "Ask Codex to do anything", "›", false},
		{"unrelated change", "abc", "abd", false},
		{"newline ignored equal base", "a\nb", "a b", true}, // after strip: "ab" vs "a b"
		{"newlines only differ then space", "a\nb", "ab ", true},
		{"non-space insert", "ab", "aXb", false},
		{"padding space run", "❯   │", "❯    │", false},
		{"boxed empty pad", " │ ❯      │", " │ ❯       │", false},
		{"space after glyph", "│ ❯", "│ ❯ ", false},
		{"draft after glyph line", "│ ❯ hello", "│ ❯ hello ", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExactlyOneMoreSpace([]byte(tc.before), []byte(tc.after))
			if got != tc.want {
				t.Fatalf("ExactlyOneMoreSpace(%q,%q)=%v want %v", tc.before, tc.after, got, tc.want)
			}
		})
	}
}

func TestStripNewlines(t *testing.T) {
	got := string(StripNewlines([]byte("a\r\nb\nc")))
	if got != "abc" {
		t.Fatalf("StripNewlines=%q want abc", got)
	}
}
