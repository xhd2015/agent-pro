package mockconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Config represents the JSON configuration file.
type Config struct {
	Port      int        `json:"port"`
	Exchanges []Exchange `json:"exchanges"`
}

// Exchange represents a request→response mapping.
type Exchange struct {
	Request  ExchangeRequest  `json:"request"`
	Response ExchangeResponse `json:"response"`
}

// ExchangeRequest defines the matching criteria.
type ExchangeRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Index   int    `json:"index"`
}

// ExchangeResponse defines the response to return.
type ExchangeResponse struct {
	Content      *string    `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls"`
	FinishReason string     `json:"finish_reason"`
}

// ToolCall represents an OpenAI tool call in the config DSL.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a function call within a tool call.
type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ParsedExchange is an exchange with metadata for matching.
type ParsedExchange struct {
	Exchange Exchange
	HasIndex bool
}

// Loaded holds a parsed config ready for the mock server.
type Loaded struct {
	Config           Config
	Exchanges        []ParsedExchange
	EffectiveIndices []int
}

// DefaultConfigJSON is used when no config path or env is provided.
const DefaultConfigJSON = `{"exchanges":[]}`

// ResolveConfigPath returns the config path using priority:
// flagValue > LLM_MOCK_CONFIG_FILE > LLM_MOCK_CONFIG > empty (default config).
func ResolveConfigPath(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if v := os.Getenv("LLM_MOCK_CONFIG_FILE"); v != "" {
		return v, nil
	}
	if v := os.Getenv("LLM_MOCK_CONFIG"); v != "" {
		return v, nil
	}
	return "", nil
}

// LoadMerged resolves the config path and loads config with optional events merge.
func LoadMerged(flagValue string) (*Loaded, error) {
	path, err := ResolveConfigPath(flagValue)
	if err != nil {
		return nil, err
	}
	return LoadFromPath(path, os.Getenv("LLM_MOCK_EVENTS_FILE"))
}

// LoadFromPath reads config JSON from path and merges optional events JSONL.
// An empty configPath uses DefaultConfigJSON.
func LoadFromPath(configPath, eventsPath string) (*Loaded, error) {
	var data []byte
	if configPath == "" {
		data = []byte(DefaultConfigJSON)
	} else {
		var err error
		data, err = os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	return ParseAndMerge(data, eventsPath)
}

// ParseAndMerge parses config JSON and appends exchanges from events JSONL.
func ParseAndMerge(configData []byte, eventsPath string) (*Loaded, error) {
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.Port == 0 {
		config.Port = 8080
	}

	parsed, err := parseExchanges(configData, config.Exchanges)
	if err != nil {
		return nil, err
	}

	if eventsPath != "" {
		extra, err := loadEventsJSONL(eventsPath)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, extra...)
		config.Exchanges = append(config.Exchanges, exchangesFromParsed(extra)...)
	}

	effective, err := computeEffectiveIndices(parsed)
	if err != nil {
		return nil, err
	}

	return &Loaded{
		Config:           config,
		Exchanges:        parsed,
		EffectiveIndices: effective,
	}, nil
}

// MarshalJSON returns the loaded config as JSON bytes.
func (l *Loaded) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Config)
}

func parseExchanges(configData []byte, exchanges []Exchange) ([]ParsedExchange, error) {
	var rawCfg struct {
		Exchanges []struct {
			Request json.RawMessage `json:"request"`
		} `json:"exchanges"`
	}
	if err := json.Unmarshal(configData, &rawCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config exchanges: %w", err)
	}

	out := make([]ParsedExchange, 0, len(exchanges))
	for i, ex := range exchanges {
		hasIndex := false
		if i < len(rawCfg.Exchanges) {
			var reqMap map[string]any
			if err := json.Unmarshal(rawCfg.Exchanges[i].Request, &reqMap); err != nil {
				return nil, fmt.Errorf("failed to parse exchange %d request: %w", i, err)
			}
			_, hasIndex = reqMap["index"]
		}
		if !hasIndex {
			ex.Request.Index = -1
		}
		out = append(out, ParsedExchange{Exchange: ex, HasIndex: hasIndex})
	}
	return out, nil
}

func loadEventsJSONL(path string) ([]ParsedExchange, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to stat events file: %w", err)
	}
	if info.Size() == 0 {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open events file: %w", err)
	}
	defer f.Close()

	var out []ParsedExchange
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw struct {
			Request json.RawMessage `json:"request"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("failed to parse events file line %d: %w", lineNo, err)
		}

		var ex Exchange
		if err := json.Unmarshal([]byte(line), &ex); err != nil {
			return nil, fmt.Errorf("failed to parse events file line %d: %w", lineNo, err)
		}

		hasIndex := false
		var reqMap map[string]any
		if err := json.Unmarshal(raw.Request, &reqMap); err == nil {
			_, hasIndex = reqMap["index"]
		}
		if !hasIndex {
			ex.Request.Index = -1
		}
		out = append(out, ParsedExchange{Exchange: ex, HasIndex: hasIndex})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read events file: %w", err)
	}
	return out, nil
}

func exchangesFromParsed(parsed []ParsedExchange) []Exchange {
	out := make([]Exchange, len(parsed))
	for i, p := range parsed {
		out[i] = p.Exchange
	}
	return out
}

func computeEffectiveIndices(parsed []ParsedExchange) ([]int, error) {
	seen := make(map[int]struct{})
	effective := make([]int, len(parsed))
	nextIdx := 0

	for i, p := range parsed {
		idx := p.Exchange.Request.Index
		if p.HasIndex && idx >= 0 {
			if _, dup := seen[idx]; dup {
				return nil, fmt.Errorf("duplicate explicit index %d across merged exchanges", idx)
			}
			seen[idx] = struct{}{}
			effective[i] = idx
			if idx >= nextIdx {
				nextIdx = idx + 1
			}
			continue
		}
		effective[i] = nextIdx
		nextIdx++
	}
	return effective, nil
}