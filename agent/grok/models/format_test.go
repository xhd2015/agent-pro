package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatTextMarksDefault(t *testing.T) {
	t.Parallel()
	cat := Catalog{
		Home:    "/tmp/.grok",
		Default: "grok-4.5",
		Models: []Model{
			{ID: "ais-glm-5-2", Source: DefaultConfigFile, DisplayName: "AIS - GLM-5.2"},
			{ID: "grok-4.5", Source: DefaultConfigFile, DisplayName: "Grok 4.5"},
			{ID: "grok-4.6", Source: ModelsCacheFile},
		},
	}
	out := FormatText(cat)
	if !strings.Contains(out, "* grok-4.5  Grok 4.5\n") {
		t.Fatalf("missing default mark:\n%s", out)
	}
	if !strings.Contains(out, "  ais-glm-5-2  AIS - GLM-5.2\n") || !strings.Contains(out, "  grok-4.6\n") {
		t.Fatalf("missing indented models:\n%s", out)
	}
	if strings.Contains(out, "* ais-glm-5-2") {
		t.Fatalf("non-default marked:\n%s", out)
	}
}

func TestFormatTextEmpty(t *testing.T) {
	t.Parallel()
	out := FormatText(Catalog{Home: "/tmp/.grok", Models: []Model{}})
	if !strings.Contains(out, "(no models)") {
		t.Fatalf("output=%q", out)
	}
}

func TestFormatJSONRoundTrip(t *testing.T) {
	t.Parallel()
	cat := Catalog{
		Home:    "/tmp/.grok",
		Default: "grok-4.5",
		Models: []Model{
			{ID: "grok-4.5", Source: DefaultConfigFile, DisplayName: "Grok 4.5"},
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
	if m.ID != "grok-4.5" || m.Source != DefaultConfigFile || m.DisplayName != "Grok 4.5" {
		t.Fatalf("model=%+v", m)
	}
	if strings.Contains(string(raw), `"slug"`) {
		t.Fatalf("unexpected slug key in json:\n%s", raw)
	}
}
