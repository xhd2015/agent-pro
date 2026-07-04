package tty

import "testing"

func TestCommandExtraPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		argv []string
		want []string
	}{
		{
			name: "npm shim absolute path",
			argv: []string{"/Users/me/installed/node-v24.10.0-darwin-arm64/bin/codex", "--flag"},
			want: []string{"/Users/me/installed/node-v24.10.0-darwin-arm64/bin"},
		},
		{
			name: "bare name",
			argv: []string{"codex"},
			want: nil,
		},
		{
			name: "empty",
			argv: nil,
			want: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := commandExtraPaths(tc.argv)
			if len(got) != len(tc.want) {
				t.Fatalf("commandExtraPaths() = %#v, want %#v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("commandExtraPaths()[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}