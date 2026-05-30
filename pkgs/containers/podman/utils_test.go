package podman

import (
	"testing"
)

func TestFilterEnv(t *testing.T) {
	tests := []struct {
		name   string
		env    []string
		keys   []string
		expect []string
	}{
		{
			name:   "remove single key",
			env:    []string{"PATH=/usr/bin", "GOFLAGS=-mod=vendor", "HOME=/root"},
			keys:   []string{"GOFLAGS"},
			expect: []string{"PATH=/usr/bin", "HOME=/root"},
		},
		{
			name:   "remove multiple keys",
			env:    []string{"A=1", "B=2", "C=3"},
			keys:   []string{"A", "C"},
			expect: []string{"B=2"},
		},
		{
			name:   "no keys removes nothing",
			env:    []string{"A=1", "B=2"},
			keys:   []string{},
			expect: []string{"A=1", "B=2"},
		},
		{
			name:   "empty env",
			env:    nil,
			keys:   []string{"A"},
			expect: nil,
		},
		{
			name:   "key with empty value preserved",
			env:    []string{"A="},
			keys:   []string{"B"},
			expect: []string{"A="},
		},
		{
			name:   "prefix match only (A should not match AB)",
			env:    []string{"AB=1"},
			keys:   []string{"A"},
			expect: []string{"AB=1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterEnv(tt.env, tt.keys...)
			if len(got) != len(tt.expect) {
				t.Errorf("FilterEnv() len = %d, want %d; got=%v want=%v",
					len(got), len(tt.expect), got, tt.expect)
				return
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("FilterEnv()[%d] = %q, want %q", i, got[i], tt.expect[i])
				}
			}
		})
	}
}

func TestConfigHash(t *testing.T) {
	h1 := ConfigHash("hello")
	h2 := ConfigHash("hello")
	h3 := ConfigHash("world")

	if len(h1) != 16 {
		t.Errorf("ConfigHash() length = %d, want 16 (8 bytes hex-encoded)", len(h1))
	}
	if h1 != h2 {
		t.Errorf("ConfigHash() not deterministic: %s != %s", h1, h2)
	}
	if h1 == h3 {
		t.Errorf("ConfigHash() collision for different inputs: %s == %s", h1, h3)
	}
}

func TestConfigHashEmpty(t *testing.T) {
	h := ConfigHash("")
	if len(h) != 16 {
		t.Errorf("ConfigHash(\"\") length = %d, want 16 (8 bytes hex-encoded)", len(h))
	}
}
