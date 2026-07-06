package ttywatch

import "testing"

func TestBuildSendPayloadConditionalCR(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		suffixCR bool
		want     string
	}{
		{name: "append cr when missing", message: "follow up", suffixCR: true, want: "follow up\r"},
		{name: "preserve existing cr", message: "follow up\r", suffixCR: true, want: "follow up\r"},
		{name: "append cr after newline", message: "line1\nline2", suffixCR: true, want: "line1\nline2\r"},
		{name: "no extra cr when crlf present", message: "already\r\n", suffixCR: true, want: "already\r\n"},
		{name: "verbatim when suffix disabled", message: "hello", suffixCR: false, want: "hello"},
		{name: "preserve whitespace verbatim", message: "  spaced  ", suffixCR: true, want: "  spaced  \r"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildSendPayload(tc.message, tc.suffixCR)
			if got != tc.want {
				t.Fatalf("buildSendPayload(%q, %v) = %q, want %q", tc.message, tc.suffixCR, got, tc.want)
			}
		})
	}
}