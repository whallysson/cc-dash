package index

import (
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/whallysson/cc-dash/internal/model"
)

// Index é o índice central in-memory de todas as sessões.
// Thread-safe via RWMutex.
type Index struct {
	mu sync.RWMutex

	// Mapas primários
	sessions   map[string]*model.SessionMeta // sessionID -> meta
	projects   map[string][]string           // slug -> []sessionID
	fileStates map[string]model.FileState    // filePath -> state
	fileToSess map[string]string             // filePath -> sessionID

	// Cache de aggregações (invalidado em qualquer write)
	aggDirty   bool
	aggStats   *model.OverviewStats
	aggCosts   *model.CostAnalytics
	aggTools   *model.ToolAnalytics
	aggActivity *model.ActivityData

	claudeDir string

	// Callback para persistência (set externamente)
	OnSessionUpdate func(meta *model.SessionMeta, state model.FileState)
}

// New cria um novo índice vazio.
func New(claudeDir string) *Index {
	return &Index{
		sessions:   make(map[string]*model.SessionMeta),
		projects:   make(map[string][]string),
		fileStates: make(map[string]model.FileState),
		fileToSess: make(map[string]string),
		claudeDir:  claudeDir,
		aggDirty:   true,
	}
}

// LoadFromCache popula o índice a partir de dados do SQLite cache.
func (idx *Index) LoadFromCache(sessions map[string]*model.SessionMeta, fileStates map[string]model.FileState) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for id, meta := range sessions {
		idx.sessions[id] = meta
		idx.projects[meta.Slug] = append(idx.projects[meta.Slug], id)
	}
	for path, state := range fileStates {
		idx.fileStates[path] = state
	}
	idx.aggDirty = true
}

// Build escaneia todos os projetos e popula o índice.
func (idx *Index) Build() error {
	start := time.Now()

	cached := idx.getCachedStates()
	results, err := ScanProjects(idx.claudeDir, cached)
	if err != nil {
		return err
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	var parsed, skipped, errors int
	for _, r := range results {
		if r.Err != nil {
			errors++
			continue
		}
		if r.Meta != nil {
			idx.insertLocked(r.Meta, r.State)
			parsed++
		} else {
			skipped++
		}
	}

	idx.aggDirty = true
	elapsed := time.Since(start)
	log.Printf("[index] build completo em %v: %d parseados, %d cached, %d erros, %d total",
		elapsed, parsed, skipped, errors, len(idx.sessions))

	return nil
}

// insertLocked insere ou atualiza uma sessão no índice. Deve ser chamado com mu.Lock().
func (idx *Index) insertLocked(meta *model.SessionMeta, state model.FileState) {
	sid := meta.SessionID
	if sid == "" {
		// Sem sessionID, usar path do arquivo como chave
		sid = state.Path
		meta.SessionID = sid
	}

	// Se já existe, merge tokens e contadores (para parsing incremental)
	if existing, ok := idx.sessions[sid]; ok {
		mergeSession(existing, meta)
	} else {
		idx.sessions[sid] = meta
		idx.projects[meta.Slug] = append(idx.projects[meta.Slug], sid)
	}

	idx.fileStates[state.Path] = state
	idx.fileToSess[state.Path] = sid

	// Persistir assincronamente
	if idx.OnSessionUpdate != nil {
		go idx.OnSessionUpdate(idx.sessions[sid], state)
	}
}

// mergeSession combina dados de parsing incremental com sessão existente.
func mergeSession(existing, incoming *model.SessionMeta) {
	existing.UserMsgCount += incoming.UserMsgCount
	existing.AsstMsgCount += incoming.AsstMsgCount
	existing.TotalMsgCount = existing.UserMsgCount + existing.AsstMsgCount

	for tool, count := range incoming.ToolCounts {
		existing.ToolCounts[tool] += count
	}

	for mdl, tokens := range incoming.ModelTokens {
		t := existing.ModelTokens[mdl]
		t.Add(tokens)
		existing.ModelTokens[mdl] = t
	}

	// Recalcular totais
	existing.TotalTokens = model.TokenUsage{}
	existing.EstimatedCost = 0
	for mdl, tokens := range existing.ModelTokens {
		existing.TotalTokens.Add(tokens)
		existing.EstimatedCost += model.CalculateCost(mdl, tokens)
	}

	if incoming.EndTime.After(existing.EndTime) {
		existing.EndTime = incoming.EndTime
		existing.DurationMin = existing.EndTime.Sub(existing.StartTime).Minutes()
	}

	existing.HasCompaction = existing.HasCompaction || incoming.HasCompaction
	existing.HasThinking = existing.HasThinking || incoming.HasThinking
	existing.UsesMCP = existing.UsesMCP || incoming.UsesMCP
	existing.UsesTaskAgent = existing.UsesTaskAgent || incoming.UsesTaskAgent
	existing.UsesWebSearch = existing.UsesWebSearch || incoming.UsesWebSearch
	existing.UsesWebFetch = existing.UsesWebFetch || incoming.UsesWebFetch

	if existing.Version == "" && incoming.Version != "" {
		existing.Version = incoming.Version
	}
	if existing.GitBranch == "" && incoming.GitBranch != "" {
		existing.GitBranch = incoming.GitBranch
	}
}

// UpdateFile re-parseia um arquivo específico e atualiza o índice.
func (idx *Index) UpdateFile(path string) (*model.SessionMeta, error) {
	cached := idx.getCachedStates()
	result := processFile(path, cached)
	if result.Err != nil {
		return nil, result.Err
	}
	if result.Meta == nil {
		return nil, nil
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.insertLocked(result.Meta, result.State)
	idx.aggDirty = true

	return result.Meta, nil
}

func (idx *Index) getCachedStates() map[string]model.FileState {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	states := make(map[string]model.FileState, len(idx.fileStates))
	for k, v := range idx.fileStates {
		states[k] = v
	}
	return states
}

// --- Métodos de leitura ---

// GetSession retorna uma sessão por ID.
func (idx *Index) GetSession(id string) *model.SessionMeta {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.sessions[id]
}

// GetAllSessions retorna todas as sessões ordenadas por data (mais recente primeiro).
func (idx *Index) GetAllSessions() []*model.SessionMeta {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	sessions := make([]*model.SessionMeta, 0, len(idx.sessions))
	for _, s := range idx.sessions {
		sessions = append(sessions, s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return sessions
}

// GetSessionsPaginated retorna sessões com paginação e busca.
func (idx *Index) GetSessionsPaginated(page, limit int, sortBy, query string) ([]*model.SessionMeta, int) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Filtrar
	var filtered []*model.SessionMeta
	for _, s := range idx.sessions {
		if query != "" {
			q := query
			match := containsCI(s.FirstPrompt, q) ||
				containsCI(s.Slug, q) ||
				containsCI(s.ProjectPath, q) ||
				containsCI(s.SessionID, q)
			if !match {
				continue
			}
		}
		filtered = append(filtered, s)
	}

	// Ordenar
	switch sortBy {
	case "date", "":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].StartTime.After(filtered[j].StartTime)
		})
	case "duration":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].DurationMin > filtered[j].DurationMin
		})
	case "messages":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].TotalMsgCount > filtered[j].TotalMsgCount
		})
	case "cost":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].EstimatedCost > filtered[j].EstimatedCost
		})
	case "tokens":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].TotalTokens.Total() > filtered[j].TotalTokens.Total()
		})
	}

	total := len(filtered)

	// Paginar
	start := (page - 1) * limit
	if start >= total {
		return nil, total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return filtered[start:end], total
}

// GetProjectSummaries retorna resumos de todos os projetos.
func (idx *Index) GetProjectSummaries() []model.ProjectSummary {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	summaries := make([]model.ProjectSummary, 0, len(idx.projects))

	for slug, sessionIDs := range idx.projects {
		summary := model.ProjectSummary{
			Slug:         slug,
			SessionCount: len(sessionIDs),
		}

		branchSet := make(map[string]bool)
		toolAgg := make(map[string]int)

		for _, sid := range sessionIDs {
			s := idx.sessions[sid]
			if s == nil {
				continue
			}

			if summary.ProjectPath == "" {
				summary.ProjectPath = s.ProjectPath
			}
			summary.TotalMessages += s.TotalMsgCount
			summary.TotalTokens += s.TotalTokens.Total()
			summary.TotalCost += s.EstimatedCost
			summary.TotalDuration += s.DurationMin

			if s.EndTime.After(summary.LastActive) {
				summary.LastActive = s.EndTime
			}
			if s.GitBranch != "" {
				branchSet[s.GitBranch] = true
			}
			for tool, count := range s.ToolCounts {
				toolAgg[tool] += count
			}
		}

		for b := range branchSet {
			summary.GitBranches = append(summary.GitBranches, b)
		}

		// Top 10 tools
		type tc struct {
			name  string
			count int
		}
		var tools []tc
		for name, count := range toolAgg {
			tools = append(tools, tc{name, count})
		}
		sort.Slice(tools, func(i, j int) bool { return tools[i].count > tools[j].count })
		for i, t := range tools {
			if i >= 10 {
				break
			}
			summary.TopTools = append(summary.TopTools, model.ToolCount{Name: t.name, Count: t.count})
		}

		summaries = append(summaries, summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].LastActive.After(summaries[j].LastActive)
	})

	return summaries
}

// GetProjectDetail retorna detalhes de um projeto específico.
func (idx *Index) GetProjectDetail(slug string) (*model.ProjectSummary, []*model.SessionMeta) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	sessionIDs, ok := idx.projects[slug]
	if !ok {
		return nil, nil
	}

	var sessions []*model.SessionMeta
	summary := &model.ProjectSummary{
		Slug:         slug,
		SessionCount: len(sessionIDs),
	}

	branchSet := make(map[string]bool)
	toolAgg := make(map[string]int)

	for _, sid := range sessionIDs {
		s := idx.sessions[sid]
		if s == nil {
			continue
		}
		sessions = append(sessions, s)

		if summary.ProjectPath == "" {
			summary.ProjectPath = s.ProjectPath
		}
		summary.TotalMessages += s.TotalMsgCount
		summary.TotalTokens += s.TotalTokens.Total()
		summary.TotalCost += s.EstimatedCost
		summary.TotalDuration += s.DurationMin

		if s.EndTime.After(summary.LastActive) {
			summary.LastActive = s.EndTime
		}
		if s.GitBranch != "" {
			branchSet[s.GitBranch] = true
		}
		for tool, count := range s.ToolCounts {
			toolAgg[tool] += count
		}
	}

	for b := range branchSet {
		summary.GitBranches = append(summary.GitBranches, b)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return summary, sessions
}

// GetOverviewStats retorna estatísticas agregadas para a overview.
func (idx *Index) GetOverviewStats() model.OverviewStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	stats := model.OverviewStats{
		TotalSessions:  len(idx.sessions),
		TotalProjects:  len(idx.projects),
		HourCounts:     make(map[string]int),
		ModelBreakdown: make(map[string]int64),
	}

	dailyMap := make(map[string]*model.DailyActivity)

	for _, s := range idx.sessions {
		stats.TotalMessages += s.TotalMsgCount
		stats.TotalTokens += s.TotalTokens.Total()
		stats.TotalCost += s.EstimatedCost

		// Atividade diária
		dateKey := s.StartTime.Format("2006-01-02")
		if da, ok := dailyMap[dateKey]; ok {
			da.Sessions++
			da.Messages += s.TotalMsgCount
			da.Tokens += s.TotalTokens.Total()
		} else {
			dailyMap[dateKey] = &model.DailyActivity{
				Date:     dateKey,
				Sessions: 1,
				Messages: s.TotalMsgCount,
				Tokens:   s.TotalTokens.Total(),
			}
		}

		// Horas de pico
		hour := s.StartTime.Format("15")
		stats.HourCounts[hour]++

		// Breakdown por modelo
		for mdl, tokens := range s.ModelTokens {
			stats.ModelBreakdown[mdl] += tokens.Total()
		}
	}

	// Converter mapa de atividade diária para slice ordenado
	for _, da := range dailyMap {
		stats.DailyActivity = append(stats.DailyActivity, *da)
	}
	sort.Slice(stats.DailyActivity, func(i, j int) bool {
		return stats.DailyActivity[i].Date < stats.DailyActivity[j].Date
	})

	// Sessões recentes (top 10)
	all := make([]*model.SessionMeta, 0, len(idx.sessions))
	for _, s := range idx.sessions {
		all = append(all, s)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].StartTime.After(all[j].StartTime)
	})
	for i := range min(10, len(all)) {
		stats.RecentSessions = append(stats.RecentSessions, *all[i])
	}

	return stats
}

// GetCostAnalytics retorna analytics de custo.
func (idx *Index) GetCostAnalytics() model.CostAnalytics {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	analytics := model.CostAnalytics{
		CostByModel: make(map[string]model.ModelCostDetail),
	}

	dailyCosts := make(map[string]float64)
	projectCosts := make(map[string]float64)

	var totalCacheReads, totalCacheWrites, totalInput int64

	for _, s := range idx.sessions {
		analytics.TotalCost += s.EstimatedCost

		dateKey := s.StartTime.Format("2006-01-02")
		dailyCosts[dateKey] += s.EstimatedCost

		projectCosts[s.Slug] += s.EstimatedCost

		for mdl, tokens := range s.ModelTokens {
			detail := analytics.CostByModel[mdl]
			detail.Model = mdl
			detail.Tokens.Add(tokens)
			detail.Cost += model.CalculateCost(mdl, tokens)
			detail.SessionCount++
			analytics.CostByModel[mdl] = detail

			totalCacheReads += tokens.CacheReadTokens
			totalCacheWrites += tokens.CacheWriteTokens
			totalInput += tokens.InputTokens
		}
	}

	// Custo por dia
	for date, cost := range dailyCosts {
		analytics.CostByDate = append(analytics.CostByDate, model.DailyCost{Date: date, Cost: cost})
	}
	sort.Slice(analytics.CostByDate, func(i, j int) bool {
		return analytics.CostByDate[i].Date < analytics.CostByDate[j].Date
	})

	// Custo por projeto
	for slug, cost := range projectCosts {
		analytics.CostByProject = append(analytics.CostByProject, model.ProjectCost{Slug: slug, Cost: cost})
	}
	sort.Slice(analytics.CostByProject, func(i, j int) bool {
		return analytics.CostByProject[i].Cost > analytics.CostByProject[j].Cost
	})

	// Cache efficiency
	analytics.CacheEfficiency = model.CacheEfficiency{
		TotalCacheReads:  totalCacheReads,
		TotalCacheWrites: totalCacheWrites,
		TotalInputTokens: totalInput,
	}
	if totalInput+totalCacheReads > 0 {
		analytics.CacheEfficiency.CacheHitRate = float64(totalCacheReads) / float64(totalInput+totalCacheReads) * 100
	}

	return analytics
}

// GetToolAnalytics retorna analytics de ferramentas.
func (idx *Index) GetToolAnalytics() model.ToolAnalytics {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	analytics := model.ToolAnalytics{
		ToolsByCategory: make(map[string][]model.ToolRankEntry),
	}

	toolCounts := make(map[string]int)
	toolSessions := make(map[string]int)
	versionMap := make(map[string]*model.VersionEntry)
	featureCounts := map[string]int{
		"compaction":  0,
		"thinking":    0,
		"mcp":         0,
		"task_agent":  0,
		"web_search":  0,
		"web_fetch":   0,
	}

	total := len(idx.sessions)

	for _, s := range idx.sessions {
		seen := make(map[string]bool)
		for tool, count := range s.ToolCounts {
			toolCounts[tool] += count
			if !seen[tool] {
				toolSessions[tool]++
				seen[tool] = true
			}
		}

		if s.Version != "" {
			if ve, ok := versionMap[s.Version]; ok {
				ve.SessionCount++
			} else {
				versionMap[s.Version] = &model.VersionEntry{
					Version:      s.Version,
					FirstSeen:    s.StartTime,
					SessionCount: 1,
				}
			}
		}

		if s.HasCompaction { featureCounts["compaction"]++ }
		if s.HasThinking { featureCounts["thinking"]++ }
		if s.UsesMCP { featureCounts["mcp"]++ }
		if s.UsesTaskAgent { featureCounts["task_agent"]++ }
		if s.UsesWebSearch { featureCounts["web_search"]++ }
		if s.UsesWebFetch { featureCounts["web_fetch"]++ }
	}

	// Tool ranking
	for name, count := range toolCounts {
		cat := model.GetToolCategory(name)
		entry := model.ToolRankEntry{
			Name:     name,
			Count:    count,
			Category: cat,
			Sessions: toolSessions[name],
		}
		analytics.ToolRanking = append(analytics.ToolRanking, entry)
		analytics.ToolsByCategory[cat] = append(analytics.ToolsByCategory[cat], entry)
	}
	sort.Slice(analytics.ToolRanking, func(i, j int) bool {
		return analytics.ToolRanking[i].Count > analytics.ToolRanking[j].Count
	})

	// Version history
	for _, ve := range versionMap {
		analytics.VersionHistory = append(analytics.VersionHistory, *ve)
	}
	sort.Slice(analytics.VersionHistory, func(i, j int) bool {
		return analytics.VersionHistory[i].FirstSeen.Before(analytics.VersionHistory[j].FirstSeen)
	})

	// Feature adoption
	for feature, count := range featureCounts {
		pct := 0.0
		if total > 0 {
			pct = float64(count) / float64(total) * 100
		}
		analytics.FeatureAdoption = append(analytics.FeatureAdoption, model.FeatureEntry{
			Feature:      feature,
			SessionCount: count,
			Percentage:   pct,
		})
	}
	sort.Slice(analytics.FeatureAdoption, func(i, j int) bool {
		return analytics.FeatureAdoption[i].SessionCount > analytics.FeatureAdoption[j].SessionCount
	})

	return analytics
}

// GetActivityData retorna dados de atividade para heatmap e streaks.
func (idx *Index) GetActivityData() model.ActivityData {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	data := model.ActivityData{
		DayOfWeek: make(map[string]int),
		PeakHours: make(map[string]int),
	}

	dateSet := make(map[string]int)

	for _, s := range idx.sessions {
		dateKey := s.StartTime.Format("2006-01-02")
		dateSet[dateKey]++

		dow := s.StartTime.Weekday().String()
		data.DayOfWeek[dow]++

		hour := s.StartTime.Format("15")
		data.PeakHours[hour]++
	}

	// Heatmap
	for date, count := range dateSet {
		data.Heatmap = append(data.Heatmap, model.HeatmapEntry{Date: date, Count: count})
	}
	sort.Slice(data.Heatmap, func(i, j int) bool {
		return data.Heatmap[i].Date < data.Heatmap[j].Date
	})

	data.ActiveDays = len(dateSet)

	// Calcular streaks
	if len(data.Heatmap) > 0 {
		data.CurrentStreak, data.LongestStreak = calculateStreaks(data.Heatmap)

		first, _ := time.Parse("2006-01-02", data.Heatmap[0].Date)
		last, _ := time.Parse("2006-01-02", data.Heatmap[len(data.Heatmap)-1].Date)
		data.TotalDays = int(last.Sub(first).Hours()/24) + 1
	}

	return data
}

// calculateStreaks calcula current streak e longest streak.
func calculateStreaks(heatmap []model.HeatmapEntry) (current, longest int) {
	today := time.Now().Format("2006-01-02")

	// Criar set de datas ativas
	active := make(map[string]bool)
	for _, h := range heatmap {
		active[h.Date] = true
	}

	// Current streak: contar para trás desde hoje
	d, _ := time.Parse("2006-01-02", today)
	for {
		dateStr := d.Format("2006-01-02")
		if !active[dateStr] {
			break
		}
		current++
		d = d.AddDate(0, 0, -1)
	}

	// Longest streak
	var streak int
	if len(heatmap) > 0 {
		d, _ = time.Parse("2006-01-02", heatmap[0].Date)
		last, _ := time.Parse("2006-01-02", heatmap[len(heatmap)-1].Date)

		for !d.After(last) {
			if active[d.Format("2006-01-02")] {
				streak++
				if streak > longest {
					longest = streak
				}
			} else {
				streak = 0
			}
			d = d.AddDate(0, 0, 1)
		}
	}

	return current, longest
}

// GetFilePathForSession retorna o caminho do arquivo JSONL para uma sessão.
func (idx *Index) GetFilePathForSession(sessionID string) string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	for path, sid := range idx.fileToSess {
		if sid == sessionID {
			return path
		}
	}
	return ""
}

// GetClaudeDir retorna o diretório base do Claude.
func (idx *Index) GetClaudeDir() string {
	return idx.claudeDir
}

// SessionCount retorna o número total de sessões.
func (idx *Index) SessionCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.sessions)
}

// ExportForCache exporta sessões e file states para persistência.
func (idx *Index) ExportForCache() (map[string]*model.SessionMeta, map[string]model.FileState) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	sessions := make(map[string]*model.SessionMeta, len(idx.sessions))
	for k, v := range idx.sessions {
		sessions[k] = v
	}
	states := make(map[string]model.FileState, len(idx.fileStates))
	for k, v := range idx.fileStates {
		states[k] = v
	}
	return sessions, states
}

// containsCI faz busca case-insensitive.
func containsCI(haystack, needle string) bool {
	h := strings.ToLower(haystack)
	n := strings.ToLower(needle)
	return strings.Contains(h, n)
}
