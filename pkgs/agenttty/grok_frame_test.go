package agenttty

import (
	"os"
	"strings"
	"testing"
)

func TestParseGrokFrame_recapExpandThinkingCrimeScene(t *testing.T) {
	b, err := os.ReadFile("testdata/grok-writable/grok-after_recap-expand-thinking-idle-01a03d6f.txt")
	if err != nil {
		t.Fatal(err)
	}
	f := ParseGrokFrame(string(b))
	want := []GrokSectionKind{
		GrokSectionHeader,
		GrokSectionBody,
		GrokSectionWorkedFor,
		GrokSectionRecap,
		GrokSectionComposer,
		GrokSectionHelpFooter,
	}
	got := f.kinds()
	if len(got) != len(want) {
		t.Fatalf("kinds=%v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("kinds[%d]=%q want %q (full=%v)", i, got[i], want[i], got)
		}
	}
	if !f.workedForStop() {
		t.Fatal("expected WorkedFor+stop")
	}
	if f.hasRunningIndicator() {
		t.Fatal("did not expect RunningIndicator")
	}
	busy, ok := judgeGrokFrameBusy(f)
	if !ok || busy {
		t.Fatalf("judge busy=%v ok=%v want idle", busy, ok)
	}
}

func TestParseGrokFrame_modernBusyThinking(t *testing.T) {
	b, err := os.ReadFile("testdata/grok-writable/grok-modern-busy-thinking-tasks.txt")
	if err != nil {
		t.Fatal(err)
	}
	f := ParseGrokFrame(string(b))
	if !f.hasKind(GrokSectionActivityStrip) {
		t.Fatalf("expected activity_strip, kinds=%v", f.kinds())
	}
	if f.workedForStop() {
		t.Fatal("modern busy must not have WorkedFor+stop")
	}
	busy, ok := judgeGrokFrameBusy(f)
	if !ok || !busy {
		t.Fatalf("judge busy=%v ok=%v want busy", busy, ok)
	}
}

func TestParseGrokWorkedForLine_stopHooks(t *testing.T) {
	line := "Worked for 14m52s                                                                                                                         stop  [hooks: 2]"
	meta, ok := parseGrokWorkedForLine(line)
	if !ok || !meta.Stop || meta.Duration != "14m52s" {
		t.Fatalf("meta=%+v ok=%v", meta, ok)
	}
}

func TestJudgeGrokFrameBusy_runningIndicator(t *testing.T) {
	plain := "" +
		"~/proj 1K / 10K\n" +
		"Worked for 1s stop\n" +
		"1 command still running · send a message to interrupt\n" +
		"╭---╮\n │ ❯ Build anything │\n" +
		"Ctrl+.:shortcuts\n"
	f := ParseGrokFrame(plain)
	if !f.hasRunningIndicator() {
		t.Fatalf("expected running indicator, kinds=%v", f.kinds())
	}
	busy, ok := judgeGrokFrameBusy(f)
	if !ok || !busy {
		t.Fatalf("judge busy=%v ok=%v want busy when RunningIndicator present", busy, ok)
	}
}

func TestParseGrokFrame_tipCompactModeNotComposer(t *testing.T) {
	b, err := os.ReadFile("testdata/grok-writable/grok-status-waiting-for-response-busy-w106.txt")
	if err != nil {
		t.Fatal(err)
	}
	f := ParseGrokFrame(string(b))
	if !f.hasKind(GrokSectionTip) {
		t.Fatalf("expected tip section, kinds=%v", f.kinds())
	}
	if !f.hasKind(GrokSectionComposer) {
		t.Fatalf("expected composer section, kinds=%v", f.kinds())
	}
	// Transcript "❯ list files…" must stay in body, not open composer early.
	for _, s := range f.Sections {
		if s.Kind != GrokSectionComposer {
			continue
		}
		for _, ln := range s.Lines {
			if strings.Contains(ln, "Tight on space") {
				t.Fatalf("tip line must not live in composer: %q", ln)
			}
			if strings.Contains(ln, "list files with ls -la") {
				t.Fatalf("user transcript ❯ line must not live in composer: %q", ln)
			}
		}
	}
	for _, s := range f.Sections {
		if s.Kind != GrokSectionTip {
			continue
		}
		joined := strings.Join(s.Lines, "\n")
		if !strings.Contains(joined, "Tight on space") && !strings.Contains(joined, "compact-mode") {
			t.Fatalf("tip section missing compact-mode chrome: %#v", s.Lines)
		}
	}
}

func TestParseGrokFrame_workedForNotSwallowedByTranscriptAngle(t *testing.T) {
	b, err := os.ReadFile("testdata/grok-writable/grok-status-worked-for-idle-w106.txt")
	if err != nil {
		t.Fatal(err)
	}
	f := ParseGrokFrame(string(b))
	if !f.workedForStop() {
		t.Fatalf("expected WorkedFor+stop after composer-opener fix; kinds=%v", f.kinds())
	}
	busy, ok := judgeGrokFrameBusy(f)
	if !ok || busy {
		t.Fatalf("judge busy=%v ok=%v want idle", busy, ok)
	}
}

// sectionKindsContaining returns every section kind whose lines contain substr.
// Busy chrome often also appears earlier in Body (e.g. "◆ Run Title" then status spinner).
func sectionKindsContaining(f GrokFrame, substr string) []GrokSectionKind {
	var out []GrokSectionKind
	seen := map[GrokSectionKind]bool{}
	for _, s := range f.Sections {
		for _, ln := range s.Lines {
			if strings.Contains(ln, substr) {
				if !seen[s.Kind] {
					out = append(out, s.Kind)
					seen[s.Kind] = true
				}
				break
			}
		}
	}
	return out
}

// TestParseGrokFrame_statusFixtures encodes desired ParseGrokFrame + judge
// outcomes for the status-chrome harvest. Busy chrome above the composer
// (Waiting / Responding / Thinking / tool-run spinner) must not remain only in
// Body, and must beat an earlier WorkedFor+bare-stop in the same frame.
func TestParseGrokFrame_statusFixtures(t *testing.T) {
	type tc struct {
		file string
		// Desired judge
		wantBusy bool
		wantOK   bool
		// Chrome substring that must live outside Body (dedicated status-like section).
		// Empty means no such lift is required for this fixture.
		busyChromeOutsideBody string
		// Existing section kinds that must be present.
		wantKinds []GrokSectionKind
		// Exact kind sequence when the fixture is already stable/green.
		wantKindsExact []GrokSectionKind
		// Optional: must see WorkedFor+bare-stop crumbs (idle settlers).
		wantWorkedForStop bool
		wantRunning       bool
	}
	cases := []tc{
		{
			file:              "grok-status-worked-for-idle-w106.txt",
			wantBusy:          false,
			wantOK:            true,
			wantWorkedForStop: true,
			wantKindsExact: []GrokSectionKind{
				GrokSectionHeader, GrokSectionBody, GrokSectionWorkedFor,
				GrokSectionComposer, GrokSectionHelpFooter,
			},
		},
		{
			file:              "grok-status-worked-for-idle-live-w186.txt",
			wantBusy:          false,
			wantOK:            true,
			wantWorkedForStop: true,
			wantKinds:         []GrokSectionKind{GrokSectionWorkedFor, GrokSectionRecap, GrokSectionComposer},
		},
		{
			file:        "grok-status-command-still-running-busy-live-w160.txt",
			wantBusy:    true,
			wantOK:      true,
			wantRunning: true,
			wantKindsExact: []GrokSectionKind{
				GrokSectionBody, GrokSectionRunningIndicator,
				GrokSectionComposer, GrokSectionHelpFooter,
			},
		},
		{
			file:                  "grok-status-waiting-for-response-busy-w106.txt",
			wantBusy:              true,
			wantOK:                true,
			busyChromeOutsideBody: "Waiting for response",
			wantKinds:             []GrokSectionKind{GrokSectionStatusAboveComposer, GrokSectionTip, GrokSectionComposer, GrokSectionHelpFooter},
		},
		{
			file:                  "grok-status-responding-busy-w106.txt",
			wantBusy:              true,
			wantOK:                true,
			busyChromeOutsideBody: "Responding",
			wantKinds:             []GrokSectionKind{GrokSectionStatusAboveComposer, GrokSectionComposer, GrokSectionHelpFooter},
		},
		{
			file:                  "grok-status-waiting-for-response-busy-live-w185.txt",
			wantBusy:              true,
			wantOK:                true,
			busyChromeOutsideBody: "Waiting for response",
			wantKinds:             []GrokSectionKind{GrokSectionStatusAboveComposer, GrokSectionComposer, GrokSectionHelpFooter},
		},
		{
			// Prior Worked-for+bare-stop must not idle the frame while Waiting+[stop] is visible.
			file:                  "grok-status-waiting-for-response-busy-live-w167.txt",
			wantBusy:              true,
			wantOK:                true,
			busyChromeOutsideBody: "Waiting for response",
			wantWorkedForStop:     true, // crumbs present, but status must win
			wantKinds:             []GrokSectionKind{GrokSectionWorkedFor, GrokSectionStatusAboveComposer, GrokSectionComposer},
		},
		{
			file:                  "grok-status-thinking-busy-live-w188.txt",
			wantBusy:              true,
			wantOK:                true,
			busyChromeOutsideBody: "Thinking",
			wantKinds:             []GrokSectionKind{GrokSectionStatusAboveComposer, GrokSectionComposer, GrokSectionHelpFooter},
		},
		{
			file:                  "grok-status-running-tool-busy-live-w187.txt",
			wantBusy:              true,
			wantOK:                true,
			busyChromeOutsideBody: "Clear doctest cache and re-run for RED",
			wantKinds:             []GrokSectionKind{GrokSectionStatusAboveComposer, GrokSectionComposer, GrokSectionHelpFooter},
		},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			b, err := os.ReadFile("testdata/grok-writable/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			f := ParseGrokFrame(string(b))
			t.Logf("kinds=%v", f.kinds())

			for _, k := range tc.wantKinds {
				if !f.hasKind(k) {
					t.Errorf("missing kind %q; kinds=%v", k, f.kinds())
				}
			}
			if len(tc.wantKindsExact) > 0 {
				got := f.kinds()
				if len(got) != len(tc.wantKindsExact) {
					t.Fatalf("kinds=%v want exact %v", got, tc.wantKindsExact)
				}
				for i := range tc.wantKindsExact {
					if got[i] != tc.wantKindsExact[i] {
						t.Fatalf("kinds[%d]=%q want %q (full=%v)", i, got[i], tc.wantKindsExact[i], got)
					}
				}
			}
			if tc.wantWorkedForStop && !f.workedForStop() {
				t.Errorf("expected WorkedFor+bare-stop crumbs; kinds=%v", f.kinds())
			}
			if tc.wantRunning && !f.hasRunningIndicator() {
				t.Errorf("expected RunningIndicator; kinds=%v", f.kinds())
			}
			if tc.busyChromeOutsideBody != "" {
				kindsHit := sectionKindsContaining(f, tc.busyChromeOutsideBody)
				if len(kindsHit) == 0 {
					t.Fatalf("chrome %q not found in any section; kinds=%v", tc.busyChromeOutsideBody, f.kinds())
				}
				outsideBody := false
				for _, k := range kindsHit {
					if k != GrokSectionBody {
						outsideBody = true
						break
					}
				}
				if !outsideBody {
					t.Errorf("chrome %q only in body %v; want it also in status_above_composer (or running_indicator); frame kinds=%v",
						tc.busyChromeOutsideBody, kindsHit, f.kinds())
				}
			}

			busy, ok := judgeGrokFrameBusy(f)
			if ok != tc.wantOK || busy != tc.wantBusy {
				t.Errorf("judge busy=%v ok=%v want busy=%v ok=%v (kinds=%v)", busy, ok, tc.wantBusy, tc.wantOK, f.kinds())
			}
		})
	}
}
