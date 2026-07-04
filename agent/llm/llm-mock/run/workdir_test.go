package run

import "testing"

func TestDenormalizePrivatePath(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"/private/var/folders/foo/work", "/var/folders/foo/work"},
		{"/private/tmp/work", "/tmp/work"},
		{"/var/folders/foo/work", "/var/folders/foo/work"},
		{"/Users/me/proj", "/Users/me/proj"},
	}
	for _, tc := range tests {
		if got := denormalizePrivatePath(tc.in); got != tc.want {
			t.Fatalf("denormalizePrivatePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}