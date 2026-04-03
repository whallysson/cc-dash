package index

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchCallback is called when a JSONL file changes.
type WatchCallback func(sessionID string)

// Watcher monitors ~/.claude/projects/ for changes in JSONL files.
type Watcher struct {
	idx      *Index
	watcher  *fsnotify.Watcher
	callback WatchCallback
	mu       sync.Mutex
	debounce map[string]time.Time // debounce per file
	done     chan struct{}
}

// NewWatcher creates a watcher for the index.
func NewWatcher(idx *Index, cb WatchCallback) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		idx:      idx,
		watcher:  fsw,
		callback: cb,
		debounce: make(map[string]time.Time),
		done:     make(chan struct{}),
	}

	return w, nil
}

// Start begins the watcher. Blocks until Stop() is called.
func (w *Watcher) Start() error {
	claudeDir := w.idx.GetClaudeDir()
	projectsDir := filepath.Join(claudeDir, "projects")

	// Watch project directories (not individual files - kqueue limit)
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return err
	}

	watched := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(projectsDir, entry.Name())
		if err := w.watcher.Add(dir); err != nil {
			log.Printf("[watcher] error watching %s: %v", dir, err)
			continue
		}
		watched++

		// Watch session subdirectories (subagents)
		subEntries, _ := os.ReadDir(dir)
		for _, se := range subEntries {
			if !se.IsDir() {
				continue
			}
			subDir := filepath.Join(dir, se.Name())
			subagentsDir := filepath.Join(subDir, "subagents")
			if info, err := os.Stat(subagentsDir); err == nil && info.IsDir() {
				_ = w.watcher.Add(subagentsDir)
			}
		}
	}

	// Watch history.jsonl e stats-cache.json
	_ = w.watcher.Add(claudeDir)

	log.Printf("[watcher] watching %d project directories", watched)

	go w.loop()
	return nil
}

// Stop stops the watcher.
func (w *Watcher) Stop() {
	close(w.done)
	w.watcher.Close()
}

func (w *Watcher) loop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[watcher] error: %v", err)

		case <-w.done:
			return
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Only interested in Write and Create
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
		return
	}

	path := event.Name

	// Only process JSONL files (not .summary.jsonl)
	if !strings.HasSuffix(path, ".jsonl") || strings.HasSuffix(path, ".summary.jsonl") {
		return
	}

	// Debounce: ignore repeated events within <500ms
	w.mu.Lock()
	last, exists := w.debounce[path]
	now := time.Now()
	if exists && now.Sub(last) < 500*time.Millisecond {
		w.mu.Unlock()
		return
	}
	w.debounce[path] = now
	w.mu.Unlock()

	// Re-parse the file and update the index
	go func() {
		meta, err := w.idx.UpdateFile(path)
		if err != nil {
			log.Printf("[watcher] error updating %s: %v", filepath.Base(path), err)
			return
		}
		if meta != nil && w.callback != nil {
			w.callback(meta.SessionID)
		}
	}()
}
