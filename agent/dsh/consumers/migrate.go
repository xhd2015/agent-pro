package consumers

import (
	"os"
	"path/filepath"
	"sort"
)

// RecordFiles lists the cache records a bus keeps in dir.
func RecordFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(files)
	return files, nil
}

// DropConsumer removes one consumer's cached state from every record in dir, or
// every consumer's state when id is empty.
//
// Dropping the state is how a fold whose rules changed is rebuilt: the next Sync
// sees an uncovered record, recreates the consumer at the stream origin, and
// rewinds the stream so it can replay, while the consumers whose state survived
// skip the rows they already folded. The alternative -- recomputing the fold
// here -- would be a second copy of rules this package does not own.
//
// dryRun reports what would change without writing. It returns the number of
// records scanned and the number that changed.
func DropConsumer(dir, consumerID string, dryRun bool) (scanned, changed int, err error) {
	files, err := RecordFiles(dir)
	if err != nil {
		return 0, 0, err
	}
	store := recordStore{dir: dir}
	for _, path := range files {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		var rec record
		if jsonUnmarshal(data, &rec) != nil || rec.Schema != recordSchema || rec.SessionID == "" {
			continue
		}
		scanned++
		if consumerID != "" {
			if _, ok := rec.Consumers[consumerID]; !ok {
				continue
			}
			delete(rec.Consumers, consumerID)
		} else {
			if len(rec.Consumers) == 0 {
				continue
			}
			rec.Consumers = nil
		}
		changed++
		if dryRun {
			continue
		}
		if saveErr := store.save(&rec); saveErr != nil {
			return scanned, changed, saveErr
		}
	}
	return scanned, changed, nil
}
