package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/whallysson/cc-dash/internal/model"
	"github.com/whallysson/cc-dash/internal/replay"
)

// --- Helpers ---

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

// --- Handlers ---

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.idx.GetOverviewStats()

	cache, err := model.ReadStatsCache(s.idx.GetClaudeDir())
	if err == nil && cache != nil {
		if len(cache.DailyActivity) > len(stats.DailyActivity) {
			stats.DailyActivity = cache.DailyActivity
		}
		if len(cache.HourCounts) > len(stats.HourCounts) {
			stats.HourCounts = cache.HourCounts
		}
	}

	stats.StorageBytes = model.GetClaudeStorageBytes(s.idx.GetClaudeDir())

	writeJSON(w, stats)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 50)
	sortBy := r.URL.Query().Get("sort")
	query := r.URL.Query().Get("q")

	sessions, total := s.idx.GetSessionsPaginated(page, limit, sortBy, query)

	writeJSON(w, map[string]interface{}{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (s *Server) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	session := s.idx.GetSession(id)
	if session == nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, session)
}

func (s *Server) handleSessionReplay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	offset := queryInt(r, "offset", 0)
	limit := queryInt(r, "limit", 0)

	filePath := s.idx.GetFilePathForSession(id)
	if filePath == "" {
		writeError(w, http.StatusNotFound, "session file not found")
		return
	}

	data, err := replay.ParseReplay(filePath, offset, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to parse replay: "+err.Error())
		return
	}

	writeJSON(w, data)
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	summaries := s.idx.GetProjectSummaries()
	writeJSON(w, summaries)
}

func (s *Server) handleProjectDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	summary, sessions := s.idx.GetProjectDetail(slug)
	if summary == nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	writeJSON(w, map[string]interface{}{
		"project":  summary,
		"sessions": sessions,
	})
}

func (s *Server) handleCosts(w http.ResponseWriter, r *http.Request) {
	analytics := s.idx.GetCostAnalytics()
	writeJSON(w, analytics)
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	analytics := s.idx.GetToolAnalytics()
	writeJSON(w, analytics)
}

func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	data := s.idx.GetActivityData()
	writeJSON(w, data)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 200)
	query := r.URL.Query().Get("q")

	entries, err := model.ReadHistory(s.idx.GetClaudeDir(), limit, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read history: "+err.Error())
		return
	}

	writeJSON(w, entries)
}

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	memories, err := model.ReadMemories(s.idx.GetClaudeDir())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read memories: "+err.Error())
		return
	}
	writeJSON(w, memories)
}

func (s *Server) handleMemoryUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.FilePath == "" || body.Content == "" {
		writeError(w, http.StatusBadRequest, "file_path and content are required")
		return
	}

	if err := os.WriteFile(body.FilePath, []byte(body.Content), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save: "+err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) handlePlans(w http.ResponseWriter, r *http.Request) {
	plans, err := model.ReadPlans(s.idx.GetClaudeDir())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read plans: "+err.Error())
		return
	}
	writeJSON(w, plans)
}

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := model.ReadTodos(s.idx.GetClaudeDir())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read todos: "+err.Error())
		return
	}
	writeJSON(w, todos)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	settings, _ := model.ReadSettings(s.idx.GetClaudeDir())
	plugins, _ := model.ReadInstalledPlugins(s.idx.GetClaudeDir())

	writeJSON(w, map[string]interface{}{
		"settings": settings,
		"plugins":  plugins,
	})
}

func (s *Server) handleEfficiency(w http.ResponseWriter, r *http.Request) {
	data := s.idx.GetEfficiencyData()
	writeJSON(w, data)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	sessions := s.idx.GetAllSessions()
	stats, _ := model.ReadStatsCache(s.idx.GetClaudeDir())

	// Convert pointers to values
	sessionValues := make([]model.SessionMeta, len(sessions))
	for i, s := range sessions {
		sessionValues[i] = *s
	}

	payload := model.ExportPayload{
		Version:    "1.0",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Sessions:   sessionValues,
		Stats:      stats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=cc-dash-export.json")
	json.NewEncoder(w).Encode(payload)
}
