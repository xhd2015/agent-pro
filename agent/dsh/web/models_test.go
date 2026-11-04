package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestServiceInitialModelIsPinnedAcrossCreationRetryAndLaterPreferences(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	f.failCreate = true
	opts := f.opts()
	original := ModelSelection{Provider: "route-a", Model: "same-model", ReasoningEffort: "max"}
	opts.ModelSelection = &original
	if _, err := f.service().OpenSession(context.Background(), opts); err == nil {
		t.Fatal("expected ambiguous creation failure")
	}
	// A restarted caller now has different defaults; only the persisted choice is retried.
	opts.ModelSelection = &ModelSelection{Provider: "route-b", Model: "same-model", ReasoningEffort: "low"}
	result, err := f.service().OpenSession(context.Background(), opts)
	if err != nil || !result.Accepted {
		t.Fatal(result, err)
	}
	if len(f.modelSelections) != 2 || *f.modelSelections[0] != original || *f.modelSelections[1] != original || f.creates[0] != f.creates[1] {
		t.Fatalf("creation changed route or identity: %+v %+v", f.modelSelections, f.creates)
	}
	if _, err := f.service().OpenSession(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service().Send(context.Background(), SendOptions{Key: testKey, Prompt: "followup"}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service().Open(context.Background(), testKey); err != nil {
		t.Fatal(err)
	}
	if len(f.creates) != 2 || strings.Join(f.prompts, ",") != "initial playbook,followup" {
		t.Fatal(f.creates, f.prompts)
	}
	state, _, err := f.store.Read(context.Background(), testKey)
	if err != nil || state.InitialModelSelection == nil || *state.InitialModelSelection != original {
		t.Fatal(state, err)
	}
}

func TestServiceRequiresExplicitModelAcceptanceBeforePromptOrBrowser(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		echo *ModelSelection
	}{
		{name: "old host omits model echo"},
		{name: "wrong provider", echo: &ModelSelection{Provider: "different", Model: "m", ReasoningEffort: "high"}},
		{name: "wrong effort", echo: &ModelSelection{Provider: "p", Model: "m", ReasoningEffort: "low"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f := newServiceFixture(t)
			f.noModelEcho, f.echoOverride = test.echo == nil, test.echo
			opts := f.opts()
			opts.ModelSelection = &ModelSelection{Provider: "p", Model: "m", ReasoningEffort: "high"}
			if _, err := f.service().OpenSession(context.Background(), opts); err == nil || !strings.Contains(err.Error(), "did not confirm") {
				t.Fatal(err)
			}
			state, _, _ := f.store.Read(context.Background(), testKey)
			if state.Created || len(f.prompts) != 0 || len(f.opened) != 0 {
				t.Fatal("unconfirmed model reached prompt/browser", state, f.prompts, f.opened)
			}
			f.noModelEcho, f.echoOverride = false, nil
			if result, err := f.service().OpenSession(context.Background(), opts); err != nil || !result.Accepted {
				t.Fatal(result, err)
			}
		})
	}
}

func TestServicePreservesLegacyDefaultAndPinsResolvedModelEffort(t *testing.T) {
	t.Parallel()
	for _, explicit := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy default", true: "model default effort"}[explicit], func(t *testing.T) {
			t.Parallel()
			f := newServiceFixture(t)
			opts := f.opts()
			if explicit {
				opts.ModelSelection = &ModelSelection{Provider: "p", Model: "m"}
				f.echoOverride = &ModelSelection{Provider: "p", Model: "m", ReasoningEffort: "high"}
			}
			if _, err := f.service().OpenSession(context.Background(), opts); err != nil {
				t.Fatal(err)
			}
			state, _, _ := f.store.Read(context.Background(), testKey)
			if explicit {
				if state.InitialModelSelection == nil || state.InitialModelSelection.ReasoningEffort != "high" || f.modelSelections[0].ReasoningEffort != "" {
					t.Fatal(state, f.modelSelections)
				}
			} else if state.InitialModelSelection != nil || f.modelSelections[0] != nil {
				t.Fatal("legacy caller gained an override", state, f.modelSelections)
			}
		})
	}
}

func TestInvalidInitialModelCannotReserveBinding(t *testing.T) {
	t.Parallel()
	f := newServiceFixture(t)
	opts := f.opts()
	opts.ModelSelection = &ModelSelection{Model: "without-provider"}
	if _, err := f.service().OpenSession(context.Background(), opts); err == nil {
		t.Fatal("accepted missing provider")
	}
	if _, found, err := f.store.Read(context.Background(), testKey); err != nil || found || len(f.creates) != 0 {
		t.Fatal(found, err, f.creates)
	}
}

func TestModelCatalogRetainsProviderIdentityReasoningAndPartialFailures(t *testing.T) {
	t.Parallel()
	catalog := ModelCatalog{
		Default: ModelSelection{Provider: "p", Model: "same"}, RoutableProviders: []string{"p", "other"},
		Groups: []ModelProviderGroup{
			{ID: "p", Name: "Provider", Models: []ModelCatalogModel{
				{ID: "same", Name: "Reasoner", Reasoning: &ModelReasoning{DefaultEffort: "high", Efforts: []ModelReasoningEffort{{ID: "high", Name: "High"}, {ID: "max", Name: "Max"}}}},
				{ID: "disabled", Name: "Disabled", Unavailable: "configure credentials"},
			}},
			{ID: "other", Name: "Other", Models: []ModelCatalogModel{{ID: "same", Name: "Plain"}}},
		},
		Failures: []ModelCatalogFailure{{ID: "offline", Name: "Offline", Message: "temporarily unavailable"}},
	}
	client := catalogClient(t, catalog)
	got, err := client.ModelCatalog(context.Background())
	if err != nil || !reflect.DeepEqual(got, catalog) {
		t.Fatal(got, err)
	}
}

func TestModelCatalogDistinguishesEmptyCatalogFromMalformedData(t *testing.T) {
	t.Parallel()
	for _, valid := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing catalog", true: "configured but empty"}[valid], func(t *testing.T) {
			t.Parallel()
			value := map[string]any{"default": ModelSelection{Provider: "p", Model: "m"}}
			if valid {
				value["groups"] = []ModelProviderGroup{}
			}
			_, err := catalogClient(t, value).ModelCatalog(context.Background())
			if (err == nil) != valid {
				t.Fatal(err)
			}
		})
	}
}

func catalogClient(t *testing.T, value any) Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			RPCID   string `json:"rpcId"`
			Method  string `json:"method"`
			Payload struct {
				Args map[string]any `json:"args"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.Method != "session/modelCatalog" || r.URL.Path != "/api/session/modelCatalog" || r.Header.Get("Authorization") != "Bearer catalog-token" || len(request.Payload.Args) != 0 {
			t.Errorf("unexpected catalog RPC: %+v", request)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"type": "server-response", "rpcId": request.RPCID, "result": map[string]any{"ok": true, "value": value}})
	}))
	t.Cleanup(server.Close)
	return Client{ReadListen: func() (Listen, error) { return Listen{Origin: server.URL, Token: "catalog-token"}, nil }}
}
