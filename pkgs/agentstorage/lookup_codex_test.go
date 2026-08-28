package agentstorage

import "testing"

func TestIsCodexRunner(t *testing.T) {
	t.Parallel()
	cases := []struct {
		runner string
		want   bool
	}{
		{"codex", true},
		{"codex-tty", true},
		{" codex-tty ", true},
		{"grok", false},
		{"grok-tty", false},
		{"CODEX", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsCodexRunner(tc.runner); got != tc.want {
			t.Fatalf("IsCodexRunner(%q)=%v want %v", tc.runner, got, tc.want)
		}
	}
}

func TestFindByCodexSessionID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	rsid := "019f283b-cccc-7ccc-cccc-cccccccccccc"
	if err := store.CreateSession("ar-codex", SessionMeta{
		SessionID:       "ar-codex",
		Runner:          "codex-tty",
		RunnerSessionID: rsid,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSession("ar-grok", SessionMeta{
		SessionID:       "ar-grok",
		Runner:          "grok-tty",
		RunnerSessionID: rsid,
	}); err != nil {
		t.Fatal(err)
	}

	meta, err := FindByCodexSessionID(store, rsid)
	if err != nil {
		t.Fatal(err)
	}
	if meta.SessionID != "ar-codex" {
		t.Fatalf("meta=%+v want ar-codex", meta)
	}

	if _, err := FindByCodexSessionID(store, "019f283b-ffff-7fff-ffff-ffffffffffff"); err == nil {
		t.Fatal("want not found")
	}
}
