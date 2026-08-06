package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	bookmarkStoreFileName = "session_bookmarks.json"
	bookmarkStoreVersion  = 1
	bookmarkRunnerGrok    = "grok"
)

// Bookmark is one durable catalog entry for a session across agent runners.
type Bookmark struct {
	AgentRunner     string
	SessionID       string
	SessionDir      string
	Title           string
	NumChatMessages int
	Tags            []string
	Description     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// BookmarkView is a catalog row with hybrid live enrich / orphan status.
type BookmarkView struct {
	Bookmark
	Orphaned bool
}

// EnrichMode controls list/show live enrich for grok rows.
// Zero value is EnrichLight (fast default: session_dir + summary.json only).
type EnrichMode int

const (
	// EnrichLight reads stored session_dir/summary.json only; never WalkDir/Find.
	EnrichLight EnrichMode = iota
	// EnrichOff returns catalog snapshot only; Orphaned=false; no FS checks.
	EnrichOff
	// EnrichHeavy tries light first, then Find if light did not obtain live data.
	EnrichHeavy
)

// PinOptions controls tag/description mutation on pin/update.
// Tags nil keeps existing tags; non-nil merges (union). Description nil keeps;
// non-nil sets (including empty string). ClearTags wipes tags before merge.
type PinOptions struct {
	Tags        []string
	Description *string
	ClearTags   bool
}

// ListFilter selects catalog rows. Runner "" = all runners; Tags are AND;
// Limit 0 = unlimited. Enrich zero = EnrichLight.
type ListFilter struct {
	Runner string
	Tags   []string
	Limit  int
	Enrich EnrichMode
}

type bookmarkStoreDoc struct {
	Version    int                `json:"version"`
	Bookmarks  []bookmarkStoreRow `json:"bookmarks"`
}

type bookmarkStoreRow struct {
	AgentRunner     string   `json:"agent_runner"`
	SessionID       string   `json:"session_id"`
	SessionDir      string   `json:"session_dir"`
	Title           string   `json:"title"`
	NumChatMessages int      `json:"num_chat_messages"`
	Tags            []string `json:"tags"`
	Description     string   `json:"description"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// BookmarkGrok pins a Grok session into the multi-runner catalog under
// agentProHome. Returns (bookmark, created, err). Unknown session → error
// containing "not found" with no store write. created=true on new entry.
func BookmarkGrok(agentProHome, grokHome, sessionID string, opts *PinOptions) (*Bookmark, bool, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, false, sessionNotFoundError(sessionID)
	}

	sess, err := Find(grokHome, sessionID)
	if err != nil {
		return nil, false, err
	}

	doc, err := loadBookmarkStore(agentProHome)
	if err != nil {
		return nil, false, err
	}

	now := time.Now().UTC()
	sessionDir := filepath.Dir(sess.Path)
	if !filepath.IsAbs(sessionDir) {
		if abs, absErr := filepath.Abs(sessionDir); absErr == nil {
			sessionDir = abs
		}
	}

	idx := findBookmarkIndex(doc.Bookmarks, bookmarkRunnerGrok, sessionID)
	created := idx < 0

	var row bookmarkStoreRow
	if created {
		row = bookmarkStoreRow{
			AgentRunner: bookmarkRunnerGrok,
			SessionID:   sessionID,
			CreatedAt:   now.Format(time.RFC3339),
		}
	} else {
		row = doc.Bookmarks[idx]
	}

	// Always refresh denormalized live fields.
	row.SessionDir = sessionDir
	row.Title = sess.Title
	row.NumChatMessages = sess.NumChatMessages
	row.UpdatedAt = now.Format(time.RFC3339)

	// Tags
	existingTags := normalizeTags(row.Tags)
	if opts != nil && opts.ClearTags {
		existingTags = nil
	}
	if opts != nil && opts.Tags != nil {
		row.Tags = normalizeTags(append(existingTags, opts.Tags...))
	} else if opts != nil && opts.ClearTags {
		row.Tags = normalizeTags(nil)
	} else if created {
		row.Tags = normalizeTags(nil)
	} else {
		row.Tags = existingTags
	}

	// Description
	if opts != nil && opts.Description != nil {
		row.Description = *opts.Description
	} else if created {
		row.Description = ""
	}

	if created {
		doc.Bookmarks = append(doc.Bookmarks, row)
	} else {
		doc.Bookmarks[idx] = row
	}

	if err := writeBookmarkStore(agentProHome, doc); err != nil {
		return nil, false, err
	}

	bm := rowToBookmark(row)
	return &bm, created, nil
}

// ListBookmarks loads the catalog with mode-controlled enrich for grok rows.
// Missing store → empty list. Corrupt store → error (not wiped).
// Enrich zero value = EnrichLight.
func ListBookmarks(agentProHome, grokHome string, filter ListFilter) ([]BookmarkView, []string, error) {
	doc, err := loadBookmarkStore(agentProHome)
	if err != nil {
		return nil, nil, err
	}
	if doc == nil || len(doc.Bookmarks) == 0 {
		return []BookmarkView{}, nil, nil
	}

	views := make([]BookmarkView, 0, len(doc.Bookmarks))
	var warnings []string
	for _, row := range doc.Bookmarks {
		view, warn := enrichBookmarkView(row, grokHome, filter.Enrich)
		if warn != "" {
			warnings = append(warnings, warn)
		}
		if filter.Runner != "" && !strings.EqualFold(view.AgentRunner, filter.Runner) {
			continue
		}
		if !bookmarkHasAllTags(view.Tags, filter.Tags) {
			continue
		}
		views = append(views, view)
	}

	if filter.Limit > 0 && len(views) > filter.Limit {
		views = views[:filter.Limit]
	}
	return views, warnings, nil
}

// GetBookmark returns one catalog row. runner "" requires a unique session_id
// match across runners. Enrich same modes as ListBookmarks (zero = EnrichLight).
func GetBookmark(agentProHome, runner, sessionID, grokHome string, enrich EnrichMode) (*BookmarkView, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, bookmarkNotFoundError(runner, sessionID)
	}

	doc, err := loadBookmarkStore(agentProHome)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, bookmarkNotFoundError(runner, sessionID)
	}

	matches := matchBookmarkRows(doc.Bookmarks, runner, sessionID)
	if len(matches) == 0 {
		return nil, bookmarkNotFoundError(runner, sessionID)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous session id %s: specify --runner (multiple bookmarks match)", sessionID)
	}

	view, _ := enrichBookmarkView(matches[0], grokHome, enrich)
	return &view, nil
}

// RemoveBookmark deletes a catalog entry. runner "" unique-match same as Get.
func RemoveBookmark(agentProHome, runner, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return bookmarkNotFoundError(runner, sessionID)
	}

	doc, err := loadBookmarkStore(agentProHome)
	if err != nil {
		return err
	}
	if doc == nil || len(doc.Bookmarks) == 0 {
		return bookmarkNotFoundError(runner, sessionID)
	}

	matches := matchBookmarkRows(doc.Bookmarks, runner, sessionID)
	if len(matches) == 0 {
		return bookmarkNotFoundError(runner, sessionID)
	}
	if len(matches) > 1 {
		return fmt.Errorf("ambiguous session id %s: specify --runner (multiple bookmarks match)", sessionID)
	}

	target := matches[0]
	kept := make([]bookmarkStoreRow, 0, len(doc.Bookmarks)-1)
	for _, row := range doc.Bookmarks {
		if row.AgentRunner == target.AgentRunner && row.SessionID == target.SessionID {
			continue
		}
		kept = append(kept, row)
	}
	doc.Bookmarks = kept
	return writeBookmarkStore(agentProHome, doc)
}

// FormatBookmarksTable renders a human-readable table of bookmark views.
func FormatBookmarksTable(views []BookmarkView) string {
	if len(views) == 0 {
		return "No bookmarks found"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-10s  %-38s  %5s  %-42s  %s\n", "RUNNER", "SESSION ID", "MSGS", "TITLE", "TAGS")
	for _, v := range views {
		title := v.Title
		if title == "" {
			title = "(untitled)"
		}
		tags := strings.Join(v.Tags, ",")
		fmt.Fprintf(
			&b,
			"%-10s  %-38s  %5d  %-42s  %s\n",
			v.AgentRunner,
			v.SessionID,
			v.NumChatMessages,
			truncateTitle(title),
			tags,
		)
	}
	return strings.TrimRight(b.String(), "\n")
}

// FormatBookmarkShow renders a single bookmark for detail display.
func FormatBookmarkShow(view *BookmarkView) string {
	if view == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Runner: %s\n", view.AgentRunner)
	fmt.Fprintf(&b, "Session: %s\n", view.SessionID)

	title := strings.TrimSpace(view.Title)
	if title == "" {
		title = "(untitled)"
	}
	fmt.Fprintf(&b, "Title: %s\n", title)
	fmt.Fprintf(&b, "Chat messages: %d\n", view.NumChatMessages)
	if view.SessionDir != "" {
		fmt.Fprintf(&b, "Session dir: %s\n", view.SessionDir)
	}
	if len(view.Tags) > 0 {
		fmt.Fprintf(&b, "Tags: %s\n", strings.Join(view.Tags, ", "))
	} else {
		fmt.Fprintf(&b, "Tags: (none)\n")
	}
	if view.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", view.Description)
	} else {
		fmt.Fprintf(&b, "Description: (none)\n")
	}
	if view.Orphaned {
		fmt.Fprintf(&b, "Orphaned: true\n")
	}
	if !view.CreatedAt.IsZero() {
		fmt.Fprintf(&b, "Created: %s\n", view.CreatedAt.UTC().Format(time.RFC3339))
	}
	if !view.UpdatedAt.IsZero() {
		fmt.Fprintf(&b, "Updated: %s\n", view.UpdatedAt.UTC().Format(time.RFC3339))
	}
	return strings.TrimRight(b.String(), "\n")
}

// FormatBookmarkJSON renders a bookmark, view, or list as indented JSON (no ANSI).
func FormatBookmarkJSON(v any) (string, error) {
	payload, err := toBookmarkJSON(v)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// --- store I/O ---

func bookmarkStorePath(agentProHome string) string {
	return filepath.Join(agentProHome, bookmarkStoreFileName)
}

// loadBookmarkStore returns (nil, nil) when the file is missing.
// Corrupt JSON returns an error without modifying the file.
func loadBookmarkStore(agentProHome string) (*bookmarkStoreDoc, error) {
	path := bookmarkStorePath(agentProHome)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &bookmarkStoreDoc{Version: bookmarkStoreVersion, Bookmarks: nil}, nil
		}
		return nil, fmt.Errorf("read bookmark store: %w", err)
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return nil, fmt.Errorf("bookmark store is empty or corrupt: %s", path)
	}

	var doc bookmarkStoreDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("bookmark store corrupt: %w", err)
	}
	if doc.Version == 0 {
		doc.Version = bookmarkStoreVersion
	}
	if doc.Bookmarks == nil {
		doc.Bookmarks = []bookmarkStoreRow{}
	}
	return &doc, nil
}

func writeBookmarkStore(agentProHome string, doc *bookmarkStoreDoc) error {
	if doc == nil {
		doc = &bookmarkStoreDoc{}
	}
	if doc.Version == 0 {
		doc.Version = bookmarkStoreVersion
	}
	if doc.Bookmarks == nil {
		doc.Bookmarks = []bookmarkStoreRow{}
	}

	if err := os.MkdirAll(agentProHome, 0o755); err != nil {
		return fmt.Errorf("mkdir agent-pro home: %w", err)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal bookmark store: %w", err)
	}
	data = append(data, '\n')

	path := bookmarkStorePath(agentProHome)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write bookmark store: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("commit bookmark store: %w", err)
	}
	return nil
}

// --- helpers ---

func findBookmarkIndex(rows []bookmarkStoreRow, runner, sessionID string) int {
	for i, row := range rows {
		if row.AgentRunner == runner && row.SessionID == sessionID {
			return i
		}
	}
	return -1
}

func matchBookmarkRows(rows []bookmarkStoreRow, runner, sessionID string) []bookmarkStoreRow {
	var out []bookmarkStoreRow
	for _, row := range rows {
		if row.SessionID != sessionID {
			continue
		}
		if runner != "" && row.AgentRunner != runner {
			continue
		}
		out = append(out, row)
	}
	return out
}

func enrichBookmarkView(row bookmarkStoreRow, grokHome string, mode EnrichMode) (BookmarkView, string) {
	bm := rowToBookmark(row)
	view := BookmarkView{Bookmark: bm, Orphaned: false}

	if !strings.EqualFold(row.AgentRunner, bookmarkRunnerGrok) {
		// Non-grok rows: return as stored (Orphaned=false in v1).
		return view, ""
	}

	if mode == EnrichOff {
		// Catalog snapshot only; Orphaned not computed; no FS checks.
		return view, ""
	}

	// Light path (default, and first stage of heavy): session_dir + summary.json only.
	liveOK, lightView, lightWarn := enrichBookmarkLight(row, view)
	if mode != EnrichHeavy {
		return lightView, lightWarn
	}

	// EnrichHeavy: if light obtained live data and is not orphaned, done.
	if liveOK && !lightView.Orphaned {
		return lightView, ""
	}

	// Else Find under grokHome; success → refresh; fail → orphan + warning.
	return enrichBookmarkHeavyFind(row, lightView, grokHome)
}

// enrichBookmarkLight refreshes from stored session_dir/summary.json only.
// Never calls Find. Empty session_dir → stored fields, Orphaned=false, liveOK=false.
// Missing/unreadable summary → keep snapshot, Orphaned=true + warning, liveOK=false.
func enrichBookmarkLight(row bookmarkStoreRow, view BookmarkView) (liveOK bool, result BookmarkView, warn string) {
	sessionDir := strings.TrimSpace(row.SessionDir)
	if sessionDir == "" {
		return false, view, ""
	}

	absDir := sessionDir
	if !filepath.IsAbs(absDir) {
		if abs, absErr := filepath.Abs(absDir); absErr == nil {
			absDir = abs
		}
	}

	summaryPath := filepath.Join(absDir, "summary.json")
	sess, ok := parseSummaryFile(summaryPath)
	if !ok {
		view.Orphaned = true
		warn = fmt.Sprintf("session %s is bookmarked but not found under GROK_HOME", row.SessionID)
		return false, view, warn
	}

	view.Title = sess.Title
	view.NumChatMessages = sess.NumChatMessages
	view.SessionDir = absDir
	view.Orphaned = false
	return true, view, ""
}

// enrichBookmarkHeavyFind recovers via Find after light failed to get live data.
func enrichBookmarkHeavyFind(row bookmarkStoreRow, view BookmarkView, grokHome string) (BookmarkView, string) {
	sess, err := Find(grokHome, row.SessionID)
	if err != nil {
		view.Orphaned = true
		warn := fmt.Sprintf("session %s is bookmarked but not found under GROK_HOME", row.SessionID)
		return view, warn
	}

	view.Title = sess.Title
	view.NumChatMessages = sess.NumChatMessages
	view.SessionDir = filepath.Dir(sess.Path)
	if !filepath.IsAbs(view.SessionDir) {
		if abs, absErr := filepath.Abs(view.SessionDir); absErr == nil {
			view.SessionDir = abs
		}
	}
	view.Orphaned = false
	return view, ""
}

func rowToBookmark(row bookmarkStoreRow) Bookmark {
	created, _ := parseTimestamp(row.CreatedAt)
	updated, _ := parseTimestamp(row.UpdatedAt)
	tags := normalizeTags(row.Tags)
	if tags == nil {
		tags = []string{}
	}
	return Bookmark{
		AgentRunner:     row.AgentRunner,
		SessionID:       row.SessionID,
		SessionDir:      row.SessionDir,
		Title:           row.Title,
		NumChatMessages: row.NumChatMessages,
		Tags:            tags,
		Description:     row.Description,
		CreatedAt:       created,
		UpdatedAt:       updated,
	}
}

func normalizeTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

func bookmarkHasAllTags(have, need []string) bool {
	if len(need) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(have))
	for _, t := range have {
		set[t] = struct{}{}
	}
	for _, t := range need {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := set[t]; !ok {
			return false
		}
	}
	return true
}

func bookmarkNotFoundError(runner, sessionID string) error {
	if runner != "" {
		return fmt.Errorf("bookmark not found: runner=%s session_id=%s", runner, sessionID)
	}
	return fmt.Errorf("bookmark not found: %s", sessionID)
}

// JSON DTO shapes (snake_case) for FormatBookmarkJSON.
type bookmarkJSON struct {
	AgentRunner     string   `json:"agent_runner"`
	SessionID       string   `json:"session_id"`
	SessionDir      string   `json:"session_dir,omitempty"`
	Title           string   `json:"title"`
	NumChatMessages int      `json:"num_chat_messages"`
	Tags            []string `json:"tags"`
	Description     string   `json:"description,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	UpdatedAt       string   `json:"updated_at,omitempty"`
	Orphaned        *bool    `json:"orphaned,omitempty"`
}

func bookmarkToJSON(bm Bookmark, orphaned *bool) bookmarkJSON {
	tags := bm.Tags
	if tags == nil {
		tags = []string{}
	}
	j := bookmarkJSON{
		AgentRunner:     bm.AgentRunner,
		SessionID:       bm.SessionID,
		SessionDir:      bm.SessionDir,
		Title:           bm.Title,
		NumChatMessages: bm.NumChatMessages,
		Tags:            tags,
		Description:     bm.Description,
		Orphaned:        orphaned,
	}
	if !bm.CreatedAt.IsZero() {
		j.CreatedAt = bm.CreatedAt.UTC().Format(time.RFC3339)
	}
	if !bm.UpdatedAt.IsZero() {
		j.UpdatedAt = bm.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return j
}

func toBookmarkJSON(v any) (any, error) {
	switch x := v.(type) {
	case nil:
		return nil, fmt.Errorf("nil bookmark json payload")
	case Bookmark:
		return bookmarkToJSON(x, nil), nil
	case *Bookmark:
		if x == nil {
			return nil, fmt.Errorf("nil bookmark")
		}
		return bookmarkToJSON(*x, nil), nil
	case BookmarkView:
		o := x.Orphaned
		return bookmarkToJSON(x.Bookmark, &o), nil
	case *BookmarkView:
		if x == nil {
			return nil, fmt.Errorf("nil bookmark view")
		}
		o := x.Orphaned
		return bookmarkToJSON(x.Bookmark, &o), nil
	case []BookmarkView:
		out := make([]bookmarkJSON, 0, len(x))
		for _, view := range x {
			o := view.Orphaned
			out = append(out, bookmarkToJSON(view.Bookmark, &o))
		}
		return out, nil
	case []Bookmark:
		out := make([]bookmarkJSON, 0, len(x))
		for _, bm := range x {
			out = append(out, bookmarkToJSON(bm, nil))
		}
		return out, nil
	default:
		// Pass-through for already-shaped values; still marshal as-is.
		return v, nil
	}
}
