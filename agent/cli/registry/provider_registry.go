package registry

import (
	"fmt"
	"strings"
)

type AgentProviderInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentProvider struct {
	ID    string
	Name  string
	Agent Agent
}

type AgentProviderRegistry struct {
	defaultID string
	ordered   []AgentProvider
	byID      map[string]AgentProvider
}

func NewAgentProviderRegistry(defaultID string, providers []AgentProvider) (*AgentProviderRegistry, error) {
	reg := &AgentProviderRegistry{
		byID: make(map[string]AgentProvider, len(providers)),
	}
	for _, provider := range providers {
		id := strings.TrimSpace(provider.ID)
		if id == "" {
			return nil, fmt.Errorf("agent provider id is required")
		}
		if provider.Agent == nil {
			return nil, fmt.Errorf("agent provider %q is missing implementation", id)
		}
		provider.ID = id
		if strings.TrimSpace(provider.Name) == "" {
			provider.Name = id
		}
		if _, exists := reg.byID[id]; exists {
			return nil, fmt.Errorf("duplicate agent provider id: %s", id)
		}
		reg.byID[id] = provider
		reg.ordered = append(reg.ordered, provider)
	}
	if len(reg.ordered) == 0 {
		return nil, fmt.Errorf("at least one agent provider is required")
	}
	if strings.TrimSpace(defaultID) == "" {
		reg.defaultID = reg.ordered[0].ID
		return reg, nil
	}
	if _, ok := reg.byID[defaultID]; !ok {
		return nil, fmt.Errorf("default agent provider not found: %s", defaultID)
	}
	reg.defaultID = defaultID
	return reg, nil
}

func (r *AgentProviderRegistry) DefaultID() string {
	if r == nil {
		return ""
	}
	return r.defaultID
}

func (r *AgentProviderRegistry) Resolve(id string) (AgentProvider, error) {
	if r == nil {
		return AgentProvider{}, fmt.Errorf("agent provider registry is not configured")
	}
	providerID := strings.TrimSpace(id)
	if providerID == "" {
		providerID = r.defaultID
	}
	provider, ok := r.byID[providerID]
	if !ok {
		return AgentProvider{}, fmt.Errorf("agent provider not found: %s", providerID)
	}
	return provider, nil
}

func (r *AgentProviderRegistry) List() []AgentProviderInfo {
	if r == nil {
		return nil
	}
	out := make([]AgentProviderInfo, 0, len(r.ordered))
	for _, provider := range r.ordered {
		out = append(out, AgentProviderInfo{
			ID:   provider.ID,
			Name: provider.Name,
		})
	}
	return out
}
