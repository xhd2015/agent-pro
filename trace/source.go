package trace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	agenttrace "github.com/xhd2015/agent-pro/agent_trace"
)

type Source interface {
	List() ([]AgentTraceSummary, error)
	Get(id string) (*AgentTraceDetail, error)
	Stop(id string) (*AgentTraceDetail, error)
	Delete(id string) error
	Describe() []string
}

type FocusSource struct {
	source        Source
	command       string
	includeLinked bool
}

func NewFocusSource(source Source, command string, includeLinked bool) Source {
	command = strings.TrimSpace(command)
	if source == nil || command == "" {
		return source
	}
	return FocusSource{source: source, command: command, includeLinked: includeLinked}
}

func (s FocusSource) List() ([]AgentTraceSummary, error) {
	summaries, err := s.source.List()
	if err != nil {
		return nil, err
	}
	return filterFocusedSummaries(summaries, s.command, s.includeLinked), nil
}

func (s FocusSource) Get(id string) (*AgentTraceDetail, error) {
	return s.source.Get(id)
}

func (s FocusSource) Stop(id string) (*AgentTraceDetail, error) {
	return s.source.Stop(id)
}

func (s FocusSource) Delete(id string) error {
	return s.source.Delete(id)
}

func (s FocusSource) Describe() []string {
	return s.source.Describe()
}

type DataDirSource struct {
	DataDir string
}

func NewDataDirSource(dataDir string) Source {
	return DataDirSource{DataDir: dataDir}
}

func (s DataDirSource) List() ([]AgentTraceSummary, error) {
	return loadAgentTraceSummaries(s.DataDir)
}

func (s DataDirSource) Get(id string) (*AgentTraceDetail, error) {
	return loadAgentTraceDetail(s.DataDir, id)
}

func (s DataDirSource) Stop(id string) (*AgentTraceDetail, error) {
	return markAgentTraceStopped(s.DataDir, id)
}

func (s DataDirSource) Delete(id string) error {
	return deleteAgentTraceSession(s.DataDir, id)
}

func (s DataDirSource) Describe() []string {
	root, err := agentTraceRoot(s.DataDir)
	if err != nil {
		return []string{s.DataDir}
	}
	return []string{root}
}

type RootSource struct {
	Root string
}

func NewRootSource(root string) Source {
	return RootSource{Root: filepath.Clean(root)}
}

func (s RootSource) List() ([]AgentTraceSummary, error) {
	return loadAgentTraceSummariesFromRoot(s.Root)
}

func (s RootSource) Get(id string) (*AgentTraceDetail, error) {
	dir, err := agentTraceDirForIDInTraceRoot(s.Root, id)
	if err != nil {
		return nil, err
	}
	return loadAgentTraceDetailFromDir(dir)
}

func (s RootSource) Stop(id string) (*AgentTraceDetail, error) {
	dir, err := agentTraceDirForIDInTraceRoot(s.Root, id)
	if err != nil {
		return nil, err
	}
	if err := markAgentTraceDirStopped(dir); err != nil {
		return nil, err
	}
	return loadAgentTraceDetailFromDir(dir)
}

func (s RootSource) Delete(id string) error {
	dir, err := agentTraceDirForIDInTraceRoot(s.Root, id)
	if err != nil {
		return err
	}
	return deleteAgentTraceDir(dir)
}

func (s RootSource) Describe() []string {
	return []string{s.Root}
}

type SessionDirSource struct {
	Dir string
}

func NewSessionDirSource(dir string) Source {
	return SessionDirSource{Dir: filepath.Clean(dir)}
}

func (s SessionDirSource) List() ([]AgentTraceSummary, error) {
	summary, err := loadAgentTraceSummaryFromDir(s.Dir, time.Now())
	if err != nil {
		return nil, err
	}
	return []AgentTraceSummary{summary}, nil
}

func (s SessionDirSource) Get(id string) (*AgentTraceDetail, error) {
	meta, err := readAgentTraceMetadataOrDefault(s.Dir)
	if err != nil {
		return nil, err
	}
	if id != meta.ID {
		return nil, fmt.Errorf("agent trace not found")
	}
	return loadAgentTraceDetailFromDir(s.Dir)
}

func (s SessionDirSource) Stop(id string) (*AgentTraceDetail, error) {
	meta, err := readAgentTraceMetadata(s.Dir)
	if err != nil {
		return nil, err
	}
	if id != meta.ID {
		return nil, fmt.Errorf("agent trace not found")
	}
	if err := markAgentTraceDirStopped(s.Dir); err != nil {
		return nil, err
	}
	return loadAgentTraceDetailFromDir(s.Dir)
}

func (s SessionDirSource) Delete(id string) error {
	meta, err := readAgentTraceMetadata(s.Dir)
	if err != nil {
		return err
	}
	if id != meta.ID {
		return fmt.Errorf("agent trace not found")
	}
	return deleteAgentTraceDir(s.Dir)
}

func (s SessionDirSource) Describe() []string {
	return []string{s.Dir}
}

type FileSource struct {
	Path string
}

func NewFileSource(path string) Source {
	return FileSource{Path: filepath.Clean(path)}
}

func (s FileSource) List() ([]AgentTraceSummary, error) {
	detail, err := s.detail()
	if err != nil {
		return nil, err
	}
	return []AgentTraceSummary{{
		AgentTraceMetadata: detail.Metadata,
		LogLineCount:       len(detail.RawLines),
	}}, nil
}

func (s FileSource) Get(id string) (*AgentTraceDetail, error) {
	detail, err := s.detail()
	if err != nil {
		return nil, err
	}
	if id != detail.Metadata.ID {
		return nil, fmt.Errorf("agent trace not found")
	}
	return detail, nil
}

func (s FileSource) Stop(string) (*AgentTraceDetail, error) {
	return nil, fmt.Errorf("trace source is read-only")
}

func (s FileSource) Delete(string) error {
	return fmt.Errorf("trace source is read-only")
}

func (s FileSource) Describe() []string {
	return []string{s.Path}
}

func (s FileSource) detail() (*AgentTraceDetail, error) {
	info, err := os.Stat(s.Path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("trace source is not a file: %s", s.Path)
	}
	lines, rawLines, err := readTraceRawLines(s.Path)
	if err != nil {
		return nil, err
	}
	createdAt := info.ModTime().Format(time.RFC3339Nano)
	meta := AgentTraceMetadata{
		ID:          syntheticFileTraceID(s.Path),
		Command:     "agent-events",
		CommandLine: filepath.Base(s.Path),
		Status:      "completed",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		LogPath:     s.Path,
	}
	return &AgentTraceDetail{
		Metadata: meta,
		Messages: agenttrace.ParseMessages(lines, meta.CreatedAt),
		RawLines: rawLines,
	}, nil
}

func syntheticFileTraceID(path string) string {
	base := strings.TrimSpace(filepath.Base(path))
	if base == "" || base == "." || base == string(filepath.Separator) {
		base = "events"
	}
	return sanitizeTraceID("file-" + strings.TrimSuffix(base, filepath.Ext(base)))
}

type CombinedSource struct {
	sources []combinedSourceEntry
}

type combinedSourceEntry struct {
	prefix string
	source Source
}

func NewCombinedSource(sources []Source) Source {
	entries := make([]combinedSourceEntry, 0, len(sources))
	for _, source := range sources {
		if source == nil {
			continue
		}
		entries = append(entries, combinedSourceEntry{
			prefix: fmt.Sprintf("s%d", len(entries)+1),
			source: source,
		})
	}
	if len(entries) == 1 {
		return entries[0].source
	}
	return CombinedSource{sources: entries}
}

func (s CombinedSource) List() ([]AgentTraceSummary, error) {
	var summaries []AgentTraceSummary
	for _, entry := range s.sources {
		next, err := entry.source.List()
		if err != nil {
			continue
		}
		for _, summary := range next {
			summary.ID = entry.externalID(summary.ID)
			summaries = append(summaries, summary)
		}
	}
	summaries = withAgentTraceRelationships(summaries)
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].CreatedAt > summaries[j].CreatedAt
	})
	return summaries, nil
}

func (s CombinedSource) Get(id string) (*AgentTraceDetail, error) {
	entry, innerID, err := s.resolve(id)
	if err != nil {
		return nil, err
	}
	detail, err := entry.source.Get(innerID)
	if err != nil {
		return nil, err
	}
	detail.Metadata.ID = entry.externalID(detail.Metadata.ID)
	if summaries, err := s.List(); err == nil {
		hydrateAgentTraceDetailRelationships(detail, summaries)
	}
	return detail, nil
}

func (s CombinedSource) Stop(id string) (*AgentTraceDetail, error) {
	entry, innerID, err := s.resolve(id)
	if err != nil {
		return nil, err
	}
	detail, err := entry.source.Stop(innerID)
	if err != nil {
		return nil, err
	}
	detail.Metadata.ID = entry.externalID(detail.Metadata.ID)
	if summaries, err := s.List(); err == nil {
		hydrateAgentTraceDetailRelationships(detail, summaries)
	}
	return detail, nil
}

func (s CombinedSource) Delete(id string) error {
	entry, innerID, err := s.resolve(id)
	if err != nil {
		return err
	}
	return entry.source.Delete(innerID)
}

func (s CombinedSource) Describe() []string {
	var out []string
	for _, entry := range s.sources {
		out = append(out, entry.source.Describe()...)
	}
	return out
}

func (s CombinedSource) resolve(id string) (combinedSourceEntry, string, error) {
	for _, entry := range s.sources {
		prefix := entry.prefix + ":"
		if strings.HasPrefix(id, prefix) {
			return entry, strings.TrimPrefix(id, prefix), nil
		}
	}
	return combinedSourceEntry{}, "", fmt.Errorf("agent trace not found")
}

func (e combinedSourceEntry) externalID(id string) string {
	return e.prefix + ":" + id
}

func SourceForPath(path string) (Source, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return NewRootSource(abs), nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return NewFileSource(abs), nil
	}
	if isAgentTraceSessionDir(abs) {
		return NewSessionDirSource(abs), nil
	}
	if isAgentTraceRoot(abs) {
		return NewRootSource(abs), nil
	}
	roots := discoverNestedTraceRoots(abs)
	if len(roots) > 0 {
		sources := make([]Source, 0, len(roots))
		for _, root := range roots {
			sources = append(sources, NewRootSource(root))
		}
		return NewCombinedSource(sources), nil
	}
	return NewRootSource(abs), nil
}

func DiscoverSources(homeDir, workDir string) ([]Source, error) {
	roots := discoverTraceRoots(homeDir, workDir)
	sources := make([]Source, 0, len(roots))
	for _, root := range roots {
		sources = append(sources, NewRootSource(root))
	}
	return sources, nil
}

func discoverTraceRoots(homeDir, workDir string) []string {
	seen := map[string]bool{}
	var roots []string
	add := func(path string) {
		if strings.TrimSpace(path) == "" {
			return
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			return
		}
		abs = filepath.Clean(abs)
		if seen[abs] {
			return
		}
		seen[abs] = true
		roots = append(roots, abs)
	}
	for _, base := range []string{homeDir, workDir} {
		add(filepath.Join(base, ".agent-traces"))
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			add(filepath.Join(base, entry.Name(), agentTracesDirName))
		}
	}
	return roots
}

func isAgentTraceSessionDir(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, traceMetaFileName)); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, traceLogFile)); err == nil {
		return true
	}
	return false
}

func isAgentTraceRoot(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if isAgentTraceSessionDir(filepath.Join(dir, entry.Name())) {
			return true
		}
	}
	return false
}

func discoverNestedTraceRoots(base string) []string {
	const maxDepth = 10
	seen := map[string]bool{}
	var roots []string
	var walk func(string, int)
	walk = func(dir string, depth int) {
		if depth > maxDepth {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				continue
			}
			path := filepath.Join(dir, name)
			if name == agentTracesDirName && isAgentTraceRoot(path) {
				clean := filepath.Clean(path)
				if !seen[clean] {
					seen[clean] = true
					roots = append(roots, clean)
				}
				continue
			}
			walk(path, depth+1)
		}
	}
	walk(filepath.Clean(base), 0)
	sort.Strings(roots)
	return roots
}

func withAgentTraceRelationships(summaries []AgentTraceSummary) []AgentTraceSummary {
	if len(summaries) == 0 {
		return summaries
	}
	idMap := map[string]string{}
	dirMap := map[string]string{}
	for _, summary := range summaries {
		idMap[summary.ID] = summary.ID
		if rawID := rawCombinedTraceID(summary.ID); rawID != summary.ID {
			idMap[rawID] = summary.ID
		}
		if dir := traceSessionDir(summary.AgentTraceMetadata); dir != "" {
			dirMap[filepath.Clean(dir)] = summary.ID
		}
	}

	out := make([]AgentTraceSummary, len(summaries))
	for i, summary := range summaries {
		summary.Children = nil
		if parent := strings.TrimSpace(summary.ParentTraceID); parent != "" {
			if external, ok := idMap[parent]; ok {
				summary.ParentTraceID = external
			}
		}
		if summary.ParentTraceID == "" {
			if parentDir := strings.TrimSpace(summary.ParentTraceDir); parentDir != "" {
				if external, ok := dirMap[filepath.Clean(parentDir)]; ok {
					summary.ParentTraceID = external
				}
			}
		}
		out[i] = summary
	}

	indexByID := map[string]int{}
	for i, summary := range out {
		indexByID[summary.ID] = i
	}
	for _, summary := range out {
		parentID := strings.TrimSpace(summary.ParentTraceID)
		if parentID == "" {
			continue
		}
		parentIndex, ok := indexByID[parentID]
		if !ok {
			continue
		}
		out[parentIndex].Children = append(out[parentIndex].Children, AgentTraceChild{
			ID:              summary.ID,
			Command:         summary.Command,
			CommandLine:     summary.CommandLine,
			Status:          summary.Status,
			AgentRunnerID:   summary.AgentRunnerID,
			Model:           summary.Model,
			CreatedAt:       summary.CreatedAt,
			DelegationID:    summary.DelegationID,
			DelegationLabel: summary.DelegationLabel,
		})
	}
	return out
}

func hydrateAgentTraceDetailRelationships(detail *AgentTraceDetail, summaries []AgentTraceSummary) {
	if detail == nil {
		return
	}
	for _, summary := range withAgentTraceRelationships(summaries) {
		if summary.ID != detail.Metadata.ID {
			continue
		}
		detail.Metadata.ParentTraceID = summary.ParentTraceID
		detail.Metadata.ParentTraceDir = summary.ParentTraceDir
		detail.Metadata.ParentSessionID = summary.ParentSessionID
		detail.Metadata.DelegationID = summary.DelegationID
		detail.Metadata.DelegationLabel = summary.DelegationLabel
		detail.Metadata.Children = summary.Children
		return
	}
}

func filterFocusedSummaries(summaries []AgentTraceSummary, command string, includeLinked bool) []AgentTraceSummary {
	summaries = withAgentTraceRelationships(summaries)
	command = strings.TrimSpace(command)
	keep := map[string]bool{}
	for _, summary := range summaries {
		if summary.Command != command {
			continue
		}
		keep[summary.ID] = true
		if !includeLinked {
			continue
		}
		if summary.ParentTraceID != "" {
			keep[summary.ParentTraceID] = true
		}
		for _, child := range summary.Children {
			keep[child.ID] = true
		}
	}
	if includeLinked {
		changed := true
		for changed {
			changed = false
			for _, summary := range summaries {
				if !keep[summary.ID] {
					continue
				}
				if summary.ParentTraceID != "" && !keep[summary.ParentTraceID] {
					keep[summary.ParentTraceID] = true
					changed = true
				}
				for _, child := range summary.Children {
					if !keep[child.ID] {
						keep[child.ID] = true
						changed = true
					}
				}
			}
		}
	}
	out := make([]AgentTraceSummary, 0, len(summaries))
	for _, summary := range summaries {
		if keep[summary.ID] {
			out = append(out, summary)
		}
	}
	return out
}

func rawCombinedTraceID(id string) string {
	if before, after, ok := strings.Cut(id, ":"); ok && strings.HasPrefix(before, "s") && after != "" {
		return after
	}
	return id
}

func traceSessionDir(meta AgentTraceMetadata) string {
	if strings.TrimSpace(meta.LogPath) != "" {
		return filepath.Dir(meta.LogPath)
	}
	if strings.TrimSpace(meta.PromptPath) != "" {
		return filepath.Dir(meta.PromptPath)
	}
	return ""
}

func sanitizeTraceID(id string) string {
	id = strings.TrimSpace(id)
	var b strings.Builder
	for _, r := range id {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-.")
	if out == "" {
		return "trace"
	}
	return out
}
