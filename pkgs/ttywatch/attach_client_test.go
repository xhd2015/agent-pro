package ttywatch

import (
	"bytes"
	"testing"
)

func TestNormalizeTTYOutput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "lf only adds cr for column reset",
			in:   "yes\n[Terminal exited]\n",
			want: "yes\r\n[Terminal exited]\r\n",
		},
		{
			name: "crlf unchanged",
			in:   "yes\r\n",
			want: "yes\r\n",
		},
		{
			name: "standalone cr preserved",
			in:   "MARKER_A\rMARKER_B\n",
			want: "MARKER_A\rMARKER_B\r\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeTTYOutput([]byte(tc.in))
			if !bytes.Equal(got, []byte(tc.want)) {
				t.Fatalf("normalizeTTYOutput(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}