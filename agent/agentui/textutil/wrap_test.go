package textutil

import "testing"

func TestWrapText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		width  int
		expect string
	}{
		{name: "no wrap needed", input: "hello world", width: 20, expect: "hello world"},
		{name: "wrap at word boundary", input: "hello world here", width: 11, expect: "hello world\nhere"},
		{name: "multiline input", input: "line one\nline two long", width: 8, expect: "line one\nline two\nlong"},
		{name: "long word forced break", input: "supercalifragilistic", width: 10, expect: "supercalif\nragilistic"},
		{name: "zero width passes through", input: "hello world", width: 0, expect: "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapText(tt.input, tt.width)
			if got != tt.expect {
				t.Errorf("WrapText(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expect)
			}
		})
	}
}

func TestIndexByte(t *testing.T) {
	if got := IndexByte([]byte("abc"), 'b'); got != 1 {
		t.Errorf("IndexByte found = %d, want 1", got)
	}
	if got := IndexByte([]byte("abc"), 'x'); got != -1 {
		t.Errorf("IndexByte missing = %d, want -1", got)
	}
	if got := IndexByte(nil, 'x'); got != -1 {
		t.Errorf("IndexByte empty = %d, want -1", got)
	}
}
