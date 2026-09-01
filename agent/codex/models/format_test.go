package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatTextMarksDefaultAndReasoning(t *testing.T) {
	t.Parallel()
	cat := Catalog{
		Home:    "/tmp/.codex",
		Default: "gpt-5.5",
		Models: []Model{
			{ID: "gpt-5.6-sol", Source: ModelsCacheFile, DisplayName: "GPT-5.6-Sol", DefaultReasoning: "medium", Reasoning: []string{"low", "medium", "ultra"}},
			{ID: "gpt-5.5", Source: ModelsCacheFile, DisplayName: "GPT-5.5", DefaultReasoning: "xhigh", Reasoning: []string{"low", "xhigh"}},
		},
	}
	out := FormatText(cat)
	if !strings.Contains(out, "* gpt-5.5") {
		t.Fatalf("missing default mark:\n%s", out)
	}
	if !strings.Contains(out, "  gpt-5.6-sol") {
		t.Fatalf("missing indented model:\n%s", out)
	}
	if !strings.Contains(out, "reasoning=[low medium ultra]") || !strings.Contains(out, "default=medium") {
		t.Fatalf("missing sol reasoning:\n%s", out)
	}
	if strings.Contains(out, "* gpt-5.6-sol") {
		t.Fatalf("non-default marked:\n%s", out)
	}
}

func TestFormatTextEmpty(t *testing.T) {
	t.Parallel()
	out := FormatText(Catalog{Home: "/tmp/.codex", Models: []Model{}})
	if !strings.Contains(out, "(no models)") {
		t.Fatalf("output=%q", out)
	}
}

func TestFormatJSONRoundTrip(t *testing.T) {
	t.Parallel()
	cat := Catalog{
		Home:    "/tmp/.codex",
		Default: "gpt-5.5",
		Models: []Model{
			{ID: "gpt-5.5", Source: ModelsCacheFile, DisplayName: "GPT-5.5", DefaultReasoning: "xhigh", Reasoning: []string{"low", "xhigh"}},
		},
		FromConfig: true,
		FromCache:  true,
	}
	raw, err := FormatJSON(cat)
	if err != nil {
		t.Fatal(err)
	}
	var got Catalog
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Home != cat.Home || got.Default != cat.Default || len(got.Models) != 1 {
		t.Fatalf("got=%+v", got)
	}
	m := got.Models[0]
	if m.ID != "gpt-5.5" || m.Source != ModelsCacheFile || m.DefaultReasoning != "xhigh" {
		t.Fatalf("model=%+v", m)
	}
	if strings.Contains(string(raw), `"slug"`) {
		t.Fatalf("unexpected slug key in json:\n%s", raw)
	}
}
