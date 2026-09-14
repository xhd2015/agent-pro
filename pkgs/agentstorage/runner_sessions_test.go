package agentstorage

import (
	"testing"
)

func TestCreateSessionSeedsRunnerSessions(t *testing.T) {
	t.Parallel()
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-seed"
	if err := store.CreateSession(sid, SessionMeta{
		Runner:          "codex-tty",
		SessionID:       sid,
		RunnerSessionID: "codex-uuid",
		Status:          "exited",
	}); err != nil {
		t.Fatal(err)
	}
	sess, err := store.GetSession(sid)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.RunnerSessions[RunnerFamilyCodex] != "codex-uuid" {
		t.Fatalf("map=%+v", sess.Meta.RunnerSessions)
	}
}

func TestActivateSessionRunner_firstVisitClearsLiveKeepsOtherFamily(t *testing.T) {
	t.Parallel()
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-switch"
	if err := store.CreateSession(sid, SessionMeta{
		Runner:          "codex-tty",
		SessionID:       sid,
		RunnerSessionID: "codex-uuid",
		Status:          "exited",
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.ActivateSessionRunner(sid, "grok-tty"); err != nil {
		t.Fatal(err)
	}
	sess, err := store.GetSession(sid)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.Runner != "grok-tty" {
		t.Fatalf("runner=%q", sess.Meta.Runner)
	}
	if sess.Meta.RunnerSessionID != "" {
		t.Fatalf("live=%q want empty", sess.Meta.RunnerSessionID)
	}
	if sess.Meta.RunnerSessions[RunnerFamilyCodex] != "codex-uuid" {
		t.Fatalf("codex slot lost: %+v", sess.Meta.RunnerSessions)
	}
}

func TestActivateSessionRunner_switchBackRestoresLive(t *testing.T) {
	t.Parallel()
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-back"
	if err := store.CreateSession(sid, SessionMeta{
		Runner:          "codex-tty",
		SessionID:       sid,
		RunnerSessionID: "codex-uuid",
		Status:          "exited",
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.ActivateSessionRunner(sid, "grok-tty"); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateSessionRunnerSessionID(sid, "grok-uuid"); err != nil {
		t.Fatal(err)
	}
	if err := store.ActivateSessionRunner(sid, "codex-tty"); err != nil {
		t.Fatal(err)
	}
	sess, err := store.GetSession(sid)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.Runner != "codex-tty" || sess.Meta.RunnerSessionID != "codex-uuid" {
		t.Fatalf("runner=%q live=%q", sess.Meta.Runner, sess.Meta.RunnerSessionID)
	}
	if sess.Meta.RunnerSessions[RunnerFamilyGrok] != "grok-uuid" {
		t.Fatalf("grok slot lost: %+v", sess.Meta.RunnerSessions)
	}
}

func TestClearSessionRunnerSessionID_familyScoped(t *testing.T) {
	t.Parallel()
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-clear"
	if err := store.CreateSession(sid, SessionMeta{
		Runner:          "grok-tty",
		SessionID:       sid,
		RunnerSessionID: "grok-uuid",
		RunnerSessions: map[string]string{
			RunnerFamilyGrok:  "grok-uuid",
			RunnerFamilyCodex: "codex-uuid",
		},
		Status: "exited",
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.ClearSessionRunnerSessionID(sid); err != nil {
		t.Fatal(err)
	}
	sess, err := store.GetSession(sid)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.RunnerSessionID != "" {
		t.Fatalf("live=%q", sess.Meta.RunnerSessionID)
	}
	if _, ok := sess.Meta.RunnerSessions[RunnerFamilyGrok]; ok {
		t.Fatalf("grok slot still present: %+v", sess.Meta.RunnerSessions)
	}
	if sess.Meta.RunnerSessions[RunnerFamilyCodex] != "codex-uuid" {
		t.Fatalf("codex slot lost: %+v", sess.Meta.RunnerSessions)
	}
}

func TestFindByCodexSessionID_whileCurrentRunnerIsGrok(t *testing.T) {
	t.Parallel()
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-lookup"
	const codexID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	const grokID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	if err := store.CreateSession(sid, SessionMeta{
		Runner:          "codex-tty",
		SessionID:       sid,
		RunnerSessionID: codexID,
		Status:          "exited",
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.ActivateSessionRunner(sid, "grok-tty"); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateSessionRunnerSessionID(sid, grokID); err != nil {
		t.Fatal(err)
	}
	got, err := FindByCodexSessionID(store, codexID)
	if err != nil {
		t.Fatalf("FindByCodexSessionID: %v", err)
	}
	if got.SessionID != sid {
		t.Fatalf("SessionID=%q", got.SessionID)
	}
	gotGrok, err := FindByGrokSessionID(store, grokID)
	if err != nil {
		t.Fatalf("FindByGrokSessionID: %v", err)
	}
	if gotGrok.SessionID != sid {
		t.Fatalf("grok SessionID=%q", gotGrok.SessionID)
	}
}

func TestActivateSessionRunner_sameFamilyAliasKeepsLive(t *testing.T) {
	t.Parallel()
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const sid = "sess-alias"
	if err := store.CreateSession(sid, SessionMeta{
		Runner:          "grok",
		SessionID:       sid,
		RunnerSessionID: "grok-uuid",
		Status:          "exited",
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.ActivateSessionRunner(sid, "grok-tty"); err != nil {
		t.Fatal(err)
	}
	sess, err := store.GetSession(sid)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Meta.Runner != "grok-tty" || sess.Meta.RunnerSessionID != "grok-uuid" {
		t.Fatalf("runner=%q live=%q", sess.Meta.Runner, sess.Meta.RunnerSessionID)
	}
}
