package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ModelSelection identifies an exact DSH provider route. Empty ReasoningEffort
// follows that model's adapter default; it is not a Codex effort alias.
type ModelSelection struct {
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoningEffort,omitempty"`
}

// ModelCatalog describes this running DSH deployment, including partial failures.
// A listed model is selectable only when Unavailable is empty. Listing does not
// create a session, open a browser, or save global/default configuration.
type ModelCatalog struct {
	Default           ModelSelection        `json:"default"`
	RoutableProviders []string              `json:"routableProviders"`
	Groups            []ModelProviderGroup  `json:"groups"`
	Failures          []ModelCatalogFailure `json:"failures"`
}

type ModelProviderGroup struct {
	ID     string              `json:"id"`
	Name   string              `json:"name"`
	Models []ModelCatalogModel `json:"models"`
}

type ModelCatalogModel struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Reasoning   *ModelReasoning `json:"reasoning,omitempty"`
	Unavailable string          `json:"unavailable,omitempty"`
}

type ModelReasoning struct {
	Efforts       []ModelReasoningEffort `json:"efforts"`
	DefaultEffort string                 `json:"defaultEffort,omitempty"`
}

type ModelReasoningEffort struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ModelCatalogFailure struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// ModelCatalog reads the live registry through the existing authenticated Web
// API. Provider failures are data, not a reason to discard healthy groups.
func (c Client) ModelCatalog(ctx context.Context) (ModelCatalog, error) {
	var catalog ModelCatalog
	if err := c.rpc(ctx, "session/modelCatalog", map[string]any{}, &catalog); err != nil {
		return ModelCatalog{}, err
	}
	if _, err := normalizeModelSelection(&catalog.Default); err != nil || catalog.Groups == nil {
		return ModelCatalog{}, fmt.Errorf("DSH returned an invalid model catalog")
	}
	seen := map[string]bool{}
	for _, group := range catalog.Groups {
		if strings.TrimSpace(group.ID) == "" || seen[group.ID] {
			return ModelCatalog{}, fmt.Errorf("DSH returned an invalid provider group")
		}
		seen[group.ID] = true
		models := map[string]bool{}
		for _, model := range group.Models {
			if strings.TrimSpace(model.ID) == "" || models[model.ID] {
				return ModelCatalog{}, fmt.Errorf("DSH returned an invalid model group")
			}
			models[model.ID] = true
		}
	}
	return catalog, nil
}

// SessionModel is the durable model-selection fold for one Session. Next is the
// pending route for the next request; LastUsed is the last consumed route.
// Both nil means a new Session that has never selected a model.
type SessionModel struct {
	Next     *ModelSelection `json:"next"`
	LastUsed *ModelSelection `json:"lastUsed"`
}

// SessionModel reads session/projections and returns only modelSelection.
func (c Client) SessionModel(ctx context.Context, sessionID string) (SessionModel, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return SessionModel{}, fmt.Errorf("DSH session is required")
	}
	var raw json.RawMessage
	if err := c.rpc(ctx, "session/projections", map[string]any{"request": map[string]any{"sessionId": sessionID}}, &raw); err != nil {
		return SessionModel{}, err
	}
	return sessionModelFromProjections(raw), nil
}

func sessionModelFromProjections(raw json.RawMessage) SessionModel {
	if len(bytes.TrimSpace(raw)) == 0 || string(raw) == "null" {
		return SessionModel{}
	}
	var envelope struct {
		Values struct {
			ModelSelection *SessionModel `json:"modelSelection"`
		} `json:"values"`
	}
	if json.Unmarshal(raw, &envelope) != nil || envelope.Values.ModelSelection == nil {
		return SessionModel{}
	}
	return *envelope.Values.ModelSelection
}

func normalizeModelSelection(selection *ModelSelection) (*ModelSelection, error) {
	if selection == nil {
		return nil, nil
	}
	next := ModelSelection{Provider: strings.TrimSpace(selection.Provider), Model: strings.TrimSpace(selection.Model), ReasoningEffort: strings.TrimSpace(selection.ReasoningEffort)}
	if next.Provider == "" || next.Model == "" {
		return nil, fmt.Errorf("DSH model selection requires both provider and model")
	}
	return &next, nil
}

// CreateOptions describes a reserved blank session. ModelSelection, if present,
// initializes only this session; it never calls session/selectModel.
type CreateOptions struct {
	Workspace      string
	SessionID      string
	ModelSelection *ModelSelection
}

// CreatedSession echoes the Host's installed initial selection when requested.
// The echo prevents an older Host silently ignoring an explicit override.
type CreatedSession struct {
	SessionID      string          `json:"sessionId"`
	ModelSelection *ModelSelection `json:"modelSelection,omitempty"`
}
