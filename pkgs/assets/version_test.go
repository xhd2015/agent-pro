package assets

import "testing"

func TestNormalizeVersion(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"0.0.70", "v0.0.70"},
		{"v0.0.70", "v0.0.70"},
		{"V0.0.70", "v0.0.70"},
		{"  0.0.70\n", "v0.0.70"},
		{"  v0.0.70  ", "v0.0.70"},
	}
	for _, tc := range cases {
		if got := NormalizeVersion(tc.in); got != tc.want {
			t.Errorf("NormalizeVersion(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestClientVersion(t *testing.T) {
	got := ClientVersion()
	if got == "" {
		t.Fatal("ClientVersion empty")
	}
	if got[0] != 'v' {
		t.Fatalf("ClientVersion=%q, want leading v", got)
	}
}

func TestAssetReleaseNames(t *testing.T) {
	names := AssetReleaseNames("0.0.70")
	want := []string{
		"agent-run_v0.0.70_frontend.tar.gz",
		"agent-pro_v0.0.70_frontend.tar.gz",
	}
	if len(names) != len(want) {
		t.Fatalf("len=%d, want %d: %v", len(names), len(want), names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("names[%d]=%q, want %q", i, names[i], want[i])
		}
	}
}

func TestAssetArchiveName(t *testing.T) {
	got := AssetArchiveName(ProductAgentPro, "v0.0.70", KindFrontend)
	want := "agent-pro_v0.0.70_frontend.tar.gz"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	got2 := AssetArchiveName(ProductAgentPro, "0.0.70", KindFrontend)
	if got2 != want {
		t.Fatalf("without v prefix: got %q, want %q", got2, want)
	}
}

func TestAssetReleaseURLPath(t *testing.T) {
	got := AssetReleaseURLPath(ProductAgentRun, "0.0.70", KindFrontend)
	want := "/v0.0.70/agent-run_v0.0.70_frontend.tar.gz"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
