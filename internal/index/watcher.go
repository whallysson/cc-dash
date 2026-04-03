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

// WatchCallback é chamado quando um arquivo JSONL muda.
type WatchCallback func(sessionID string)

// Watcher monitora ~/.claude/projects/ por mudanças em arquivos JSONL.
type Watcher struct {
	idx      *Index
	watcher  *fsnotify.Watcher
	callback WatchCallback
	mu       sync.Mutex
	debounce map[string]time.Time // debounce por arquivo
	done     chan struct{}
}

// NewWatcher cria um watcher para o índice.
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

// Start inicia o watcher. Bloqueia até Stop() ser chamado.
func (w *Watcher) Start() error {
	claudeDir := w.idx.GetClaudeDir()
	projectsDir := filepath.Join(claudeDir, "projects")

	// Watch nos diretórios de projeto (não nos arquivos individuais -- kqueue limit)
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
			log.Printf("[watcher] erro ao watch %s: %v", dir, err)
			continue
		}
		watched++

		// Watch subdiretórios de sessão (subagents)
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

	log.Printf("[watcher] monitorando %d diretórios de projeto", watched)

	go w.loop()
	return nil
}

// Stop para o watcher.
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
			log.Printf("[watcher] erro: %v", err)

		case <-w.done:
			return
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Só nos interessa Write e Create
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
		return
	}

	path := event.Name

	// Só processar arquivos JSONL (não .summary.jsonl)
	if !strings.HasSuffix(path, ".jsonl") || strings.HasSuffix(path, ".summary.jsonl") {
		return
	}

	// Debounce: ignorar eventos repetidos em <500ms
	w.mu.Lock()
	last, exists := w.debounce[path]
	now := time.Now()
	if exists && now.Sub(last) < 500*time.Millisecond {
		w.mu.Unlock()
		return
	}
	w.debounce[path] = now
	w.mu.Unlock()

	// Re-parsear o arquivo e atualizar o índice
	go func() {
		meta, err := w.idx.UpdateFile(path)
		if err != nil {
			log.Printf("[watcher] erro ao atualizar %s: %v", filepath.Base(path), err)
			return
		}
		if meta != nil && w.callback != nil {
			w.callback(meta.SessionID)
		}
	}()
}
