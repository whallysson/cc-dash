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

// ScanResult contém o resultado do scan de um arquivo.
type ScanResult struct {
	Meta   *model.SessionMeta
	State  model.FileState
	Err    error
}

// ScanProjects escaneia ~/.claude/projects/ e parseia todos os JSONLs em paralelo.
// Retorna os metadados de todas as sessões e os estados dos arquivos.
func ScanProjects(claudeDir string, cachedStates map[string]model.FileState) ([]ScanResult, error) {
	projectsDir := filepath.Join(claudeDir, "projects")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	// Coletar todos os arquivos JSONL (excluindo .summary.jsonl)
	// Estrutura: projects/<slug>/<session>.jsonl (principal)
	//            projects/<slug>/<session-uuid>/subagents/<subagent>.jsonl
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slugDir := filepath.Join(projectsDir, entry.Name())
		collectJSONLFiles(slugDir, &files)
	}

	log.Printf("[scanner] encontrados %d arquivos JSONL", len(files))

	// Worker pool concorrente
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

	// Enviar jobs
	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	// Coletar resultados
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

// processFile processa um arquivo JSONL, usando cache quando possível.
func processFile(path string, cachedStates map[string]model.FileState) ScanResult {
	info, err := os.Stat(path)
	if err != nil {
		return ScanResult{Err: err}
	}

	currentMtime := info.ModTime()
	currentSize := info.Size()

	// Verificar cache: se mtime e size não mudaram, skip
	if cached, ok := cachedStates[path]; ok {
		if cached.Mtime.Equal(currentMtime) && cached.Size == currentSize {
			return ScanResult{
				State: cached,
				// Meta será nil -- o chamador deve usar o cache do índice
			}
		}

		// Arquivo mudou: parsing incremental desde o último offset
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

	// Parse completo
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

// DiscoverMemoryFiles encontra todos os arquivos de memória.
func DiscoverMemoryFiles(claudeDir string) []string {
	var files []string

	// Memória global
	globalMemDir := filepath.Join(claudeDir, "memory")
	walkDir(globalMemDir, ".md", &files)

	// Memória por projeto
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

// collectJSONLFiles coleta recursivamente arquivos JSONL de um diretório de projeto.
// Vai até 3 níveis de profundidade para capturar subagent JSONLs.
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
			// Ignorar diretórios memory/ e tool-results/
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
