package sessions

import (
	"strings"
	"testing"
)

func TestParsePromptsArgsFirstAndHead(t *testing.T) {
	_, err := parsePromptsArgs([]string{"--first", "--head", "2"})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("err=%v", err)
	}
}

func TestParsePromptsArgsThisWindowAndSession(t *testing.T) {
	p, err := parsePromptsArgs([]string{"--this-window", "--session-id", "abc"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = resolvePromptsSingleSession(p, "", &PromptsOpts{})
	if err == nil || !strings.Contains(err.Error(), "exactly one session source") {
		t.Fatalf("err=%v", err)
	}
}

func TestParsePromptsArgsTabCurrent(t *testing.T) {
	p, err := parsePromptsArgs([]string{"--tab", "current", "--first"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Tab == nil || *p.Tab != "current" || !p.HeadSet || p.Head != 1 {
		t.Fatalf("%+v", p)
	}
}

func TestParsePromptsArgsGrepAnd(t *testing.T) {
	p, err := parsePromptsArgs([]string{"--grep", "a", "--grep", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Grep) != 2 || !p.GrepSet {
		t.Fatalf("%+v", p)
	}
}

func TestParsePromptsArgsMain(t *testing.T) {
	p, err := parsePromptsArgs([]string{"--first", "--main"})
	if err != nil {
		t.Fatal(err)
	}
	if !p.MainOnly || !p.HeadSet || p.Head != 1 {
		t.Fatalf("%+v", p)
	}
	p2, err := parsePromptsArgs([]string{"--main-agent"})
	if err != nil {
		t.Fatal(err)
	}
	if !p2.MainOnly {
		t.Fatalf("%+v", p2)
	}
}

func TestMainOnlyUsesSubAgentClass(t *testing.T) {
	main := Session{rawSessionKind: "main"}
	sub := Session{rawSessionKind: "subagent"}
	parentLinked := Session{rawSessionKind: "", parentSessionID: "parent-id"}
	if isSubAgentClass(main) {
		t.Fatal("main must not be subagent class")
	}
	if !isSubAgentClass(sub) || !isSubAgentClass(parentLinked) {
		t.Fatal("expected subagent class")
	}
}
