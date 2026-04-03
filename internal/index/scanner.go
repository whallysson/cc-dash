package index

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/whallysson/cc-dash/internal/model"
)

// ScanResult contains the result of scanning a file.
type ScanResult struct {
	Meta   *model.SessionMeta
	State  model.FileState
	Err    error
}

// ScanProjects scans ~/.claude/projects/ and parses all JSONLs in parallel.
// Returns metadata for all sessions and file states.
func ScanProjects(claudeDir string, cachedStates map[string]model.FileState) ([]ScanResult, error) {
	projectsDir := filepath.Join(claudeDir, "projects")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	// Collect all JSONL files (excluding .summary.jsonl)
	// Structure: projects/<slug>/<session>.jsonl (main)
	//            projects/<slug>/<session-uuid>/subagents/<subagent>.jsonl
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slugDir := filepath.Join(projectsDir, entry.Name())
		collectJSONLFiles(slugDir, &files)
	}

	log.Printf("[scanner] found %d JSONL files", len(files))

	// Concurrent worker pool
	numWorkers := runtime.NumCPU()
	if numWorkers > 12 {
		numWorkers = 12
	}

	jobs := make(chan string, len(files))
	results := make(chan ScanResult, len(files))

	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				result := processFile(path, cachedStates)
				results <- result
			}
		}()
	}

	// Send jobs
	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []ScanResult
	for r := range results {
		allResults = append(allResults, r)
	}

	return allResults, nil
}

// processFile processes a JSONL file, using cache when possible.
func processFile(path string, cachedStates map[string]model.FileState) ScanResult {
	info, err := os.Stat(path)
	if err != nil {
		return ScanResult{Err: err}
	}

	currentMtime := info.ModTime()
	currentSize := info.Size()

	// Check cache: skip if mtime and size unchanged
	if cached, ok := cachedStates[path]; ok {
		if cached.Mtime.Equal(currentMtime) && cached.Size == currentSize {
			return ScanResult{
				State: cached,
				// Meta will be nil - the caller should use the index cache
			}
		}

		// File changed: incremental parsing from last offset
		if currentSize > cached.Size && cached.Offset > 0 {
			meta, newOffset, err := ParseSessionFileFrom(path, cached.Offset)
			if err == nil && meta != nil {
				return ScanResult{
					Meta: meta,
					State: model.FileState{
						Path:   path,
						Mtime:  currentMtime,
						Size:   currentSize,
						Offset: newOffset,
					},
				}
			}
		}
	}

	// Full parse
	meta, finalOffset, err := ParseSessionFile(path)
	if err != nil {
		return ScanResult{Err: err}
	}

	return ScanResult{
		Meta: meta,
		State: model.FileState{
			Path:   path,
			Mtime:  currentMtime,
			Size:   currentSize,
			Offset: finalOffset,
		},
	}
}

// DiscoverMemoryFiles finds all memory files.
func DiscoverMemoryFiles(claudeDir string) []string {
	var files []string

	// Global memory
	globalMemDir := filepath.Join(claudeDir, "memory")
	walkDir(globalMemDir, ".md", &files)

	// Per-project memory
	projectsDir := filepath.Join(claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		memDir := filepath.Join(projectsDir, entry.Name(), "memory")
		walkDir(memDir, ".md", &files)
	}

	return files
}

// collectJSONLFiles recursively collects JSONL files from a project directory.
// Goes up to 3 levels deep to capture subagent JSONLs.
func collectJSONLFiles(dir string, files *[]string) {
	collectJSONLRecursive(dir, files, 0, 3)
}

func collectJSONLRecursive(dir string, files *[]string, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		path := filepath.Join(dir, name)
		if e.IsDir() {
			// Skip memory/ and tool-results/ directories
			if name == "memory" || name == "tool-results" {
				continue
			}
			collectJSONLRecursive(path, files, depth+1, maxDepth)
		} else if strings.HasSuffix(name, ".jsonl") && !strings.HasSuffix(name, ".summary.jsonl") {
			*files = append(*files, path)
		}
	}
}

func walkDir(dir, suffix string, files *[]string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			*files = append(*files, filepath.Join(dir, e.Name()))
		}
	}
}
