package index

import (
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/whallysson/cc-dash/internal/model"
)

// Index is the central in-memory index of all sessions.
// Thread-safe via RWMutex.
type Index struct {
	mu sync.RWMutex

	// Primary maps
	sessions   map[string]*model.SessionMeta // sessionID -> meta
	projects   map[string][]string           // slug -> []sessionID
	fileStates map[string]model.FileState    // filePath -> state
	fileToSess map[string]string             // filePath -> sessionID

	// Aggregation cache (invalidated on any write)
	aggDirty   bool
	aggStats   *model.OverviewStats
	aggCosts   *model.CostAnalytics
	aggTools   *model.ToolAnalytics
	aggActivity *model.ActivityData

	claudeDir string

	// Callback for persistence (set externally)
	OnSessionUpdate func(meta *model.SessionMeta, state model.FileState)
}

// New creates a new empty index.
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

// LoadFromCache populates the index from SQLite cache data.
func (idx *Index) LoadFromCache(sessions map[string]*model.SessionMeta, fileStates map[string]model.FileState) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for id, meta := range sessions {
		idx.sessions[id] = meta
		idx.projects[meta.Slug] = append(idx.projects[meta.Slug], id)
	}
	for path, state := range fileStates {
		idx.fileStates[path] = state
		if state.SessionID != "" {
			idx.fileToSess[path] = state.SessionID
		}
	}
	// Backfill fileToSess for cached file_states with empty session_id
	// (from older schema before the session_id column was added)
	for path := range idx.fileStates {
		if _, ok := idx.fileToSess[path]; ok {
			continue
		}
		for id, meta := range idx.sessions {
			if meta.SourceFile == path || strings.HasSuffix(path, id+".jsonl") {
				idx.fileToSess[path] = id
				break
			}
		}
	}
	idx.aggDirty = true
}

// Build scans all projects and populates the index.
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
	log.Printf("[index] build complete in %v: %d parsed, %d cached, %d errors, %d total",
		elapsed, parsed, skipped, errors, len(idx.sessions))

	return nil
}

// insertLocked inserts or updates a session in the index. Must be called with mu.Lock() held.
func (idx *Index) insertLocked(meta *model.SessionMeta, state model.FileState) {
	sid := meta.SessionID
	if sid == "" {
		// No sessionID, use file path as key
		sid = state.Path
		meta.SessionID = sid
	}

	// If already exists, merge tokens and counters (for incremental parsing)
	if existing, ok := idx.sessions[sid]; ok {
		mergeSession(existing, meta)
	} else {
		idx.sessions[sid] = meta
		idx.projects[meta.Slug] = append(idx.projects[meta.Slug], sid)
	}

	state.SessionID = sid
	idx.fileStates[state.Path] = state
	idx.fileToSess[state.Path] = sid

	// Persist asynchronously
	if idx.OnSessionUpdate != nil {
		go idx.OnSessionUpdate(idx.sessions[sid], state)
	}
}

// mergeSession merges incremental parsing data with an existing session.
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

	// Recalculate totals
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

// UpdateFile re-parses a specific file and updates the index.
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

// --- Read methods ---

// GetSession returns a session by ID.
func (idx *Index) GetSession(id string) *model.SessionMeta {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.sessions[id]
}

// GetAllSessions returns all sessions sorted by date (most recent first).
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

// GetSessionsPaginated returns sessions with pagination and search.
func (idx *Index) GetSessionsPaginated(page, limit int, sortBy, query string) ([]*model.SessionMeta, int) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Filter
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

	// Sort
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

	// Paginate
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

// GetProjectSummaries returns summaries of all projects.
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

// GetProjectDetail returns details of a specific project.
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

// GetOverviewStats returns aggregated statistics for the overview.
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

		// Daily activity
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

		// Peak hours
		hour := s.StartTime.Format("15")
		stats.HourCounts[hour]++

		// Breakdown by model
		for mdl, tokens := range s.ModelTokens {
			stats.ModelBreakdown[mdl] += tokens.Total()
		}
	}

	// Convert daily activity map to sorted slice
	for _, da := range dailyMap {
		stats.DailyActivity = append(stats.DailyActivity, *da)
	}
	sort.Slice(stats.DailyActivity, func(i, j int) bool {
		return stats.DailyActivity[i].Date < stats.DailyActivity[j].Date
	})

	// Recent sessions (top 10)
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

// GetCostAnalytics returns cost analytics.
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

	// Cost per day
	for date, cost := range dailyCosts {
		analytics.CostByDate = append(analytics.CostByDate, model.DailyCost{Date: date, Cost: cost})
	}
	sort.Slice(analytics.CostByDate, func(i, j int) bool {
		return analytics.CostByDate[i].Date < analytics.CostByDate[j].Date
	})

	// Cost per project
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

// GetToolAnalytics returns tool analytics.
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

// GetActivityData returns activity data for heatmap and streaks.
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

	// Calculate streaks
	if len(data.Heatmap) > 0 {
		data.CurrentStreak, data.LongestStreak = calculateStreaks(data.Heatmap)

		first, _ := time.Parse("2006-01-02", data.Heatmap[0].Date)
		last, _ := time.Parse("2006-01-02", data.Heatmap[len(data.Heatmap)-1].Date)
		data.TotalDays = int(last.Sub(first).Hours()/24) + 1
	}

	return data
}

// calculateStreaks computes the current streak and longest streak.
func calculateStreaks(heatmap []model.HeatmapEntry) (current, longest int) {
	today := time.Now().Format("2006-01-02")

	// Build set of active dates
	active := make(map[string]bool)
	for _, h := range heatmap {
		active[h.Date] = true
	}

	// Current streak: count backwards from today
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

// GetFilePathForSession returns the JSONL file path for a session.
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

// GetClaudeDir returns the Claude base directory.
func (idx *Index) GetClaudeDir() string {
	return idx.claudeDir
}

// SessionCount returns the total number of sessions.
func (idx *Index) SessionCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.sessions)
}

// GetEfficiencyData computes token health and efficiency metrics.
func (idx *Index) GetEfficiencyData() *model.EfficiencyData {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	data := &model.EfficiencyData{
		TotalSessions: len(idx.sessions),
	}

	if len(idx.sessions) == 0 {
		return data
	}

	// Collect cost-per-message for all sessions with messages
	var costPerMsgs []float64
	var totalCost float64

	// Model aggregation
	type modelAgg struct {
		sessions    int
		cost        float64
		messages    int
		tokens      int64
		input       int64
		output      int64
		cacheRead   int64
		cacheWrite  int64
	}
	models := make(map[string]*modelAgg)

	// Thinking aggregation
	var thinkSessions, noThinkSessions []float64
	var thinkCostTotal, noThinkCostTotal float64
	var thinkDurTotal, noThinkDurTotal float64
	var thinkMsgTotal, noThinkMsgTotal int

	// Vampire candidates
	type vampireCandidate struct {
		session *model.SessionMeta
		cpm     float64
	}
	var vampires []vampireCandidate

	// Cache by session
	type cacheCandidate struct {
		session  *model.SessionMeta
		hitRate  float64
		wasted   int64
	}
	var cacheCandidates []cacheCandidate

	for _, s := range idx.sessions {
		totalCost += s.EstimatedCost

		// Cost per message
		if s.TotalMsgCount > 0 {
			cpm := s.EstimatedCost / float64(s.TotalMsgCount)
			costPerMsgs = append(costPerMsgs, cpm)
		}

		// Model aggregation - prorate cost and messages by token share
		sessionTotal := max(s.TotalTokens.Total(), 1)
		for m, tu := range s.ModelTokens {
			agg, ok := models[m]
			if !ok {
				agg = &modelAgg{}
				models[m] = agg
			}
			share := float64(tu.Total()) / float64(sessionTotal)
			agg.sessions++
			agg.cost += s.EstimatedCost * share
			agg.messages += int(math.Round(float64(s.TotalMsgCount) * share))
			agg.tokens += tu.Total()
			agg.input += tu.InputTokens
			agg.output += tu.OutputTokens
			agg.cacheRead += tu.CacheReadTokens
			agg.cacheWrite += tu.CacheWriteTokens
		}

		// Thinking impact
		if s.TotalMsgCount > 0 {
			cpm := s.EstimatedCost / float64(s.TotalMsgCount)
			if s.HasThinking {
				thinkSessions = append(thinkSessions, cpm)
				thinkCostTotal += s.EstimatedCost
				thinkDurTotal += s.DurationMin
				thinkMsgTotal += s.TotalMsgCount
			} else {
				noThinkSessions = append(noThinkSessions, cpm)
				noThinkCostTotal += s.EstimatedCost
				noThinkDurTotal += s.DurationMin
				noThinkMsgTotal += s.TotalMsgCount
			}
		}

		// Vampire candidates
		if s.EstimatedCost > 0 {
			cpm := 0.0
			if s.TotalMsgCount > 0 {
				cpm = s.EstimatedCost / float64(s.TotalMsgCount)
			}
			vampires = append(vampires, vampireCandidate{session: s, cpm: cpm})
		}

		// Cache efficiency per session
		totalInput := s.TotalTokens.InputTokens + s.TotalTokens.CacheReadTokens + s.TotalTokens.CacheWriteTokens
		if totalInput > 0 {
			hitRate := float64(s.TotalTokens.CacheReadTokens) / float64(totalInput) * 100
			wasted := s.TotalTokens.CacheWriteTokens // cache writes that could have been reads
			cacheCandidates = append(cacheCandidates, cacheCandidate{
				session: s,
				hitRate: hitRate,
				wasted:  wasted,
			})
		}
	}

	data.TotalCost = totalCost

	// Cost per message stats
	if len(costPerMsgs) > 0 {
		sort.Float64s(costPerMsgs)
		n := len(costPerMsgs)
		sum := 0.0
		for _, v := range costPerMsgs {
			sum += v
		}
		data.CostPerMessage = model.CostPerMessageStats{
			Mean:   sum / float64(n),
			Median: percentile(costPerMsgs, 50),
			P90:    percentile(costPerMsgs, 90),
			P99:    percentile(costPerMsgs, 99),
			Min:    costPerMsgs[0],
			Max:    costPerMsgs[n-1],
		}
	}

	// Cost distribution histogram
	data.CostDistribution = buildCostDistribution(costPerMsgs)

	// Model comparison
	for m, agg := range models {
		totalTokens := agg.input + agg.output + agg.cacheRead + agg.cacheWrite
		cacheTotal := agg.input + agg.cacheRead + agg.cacheWrite
		hitRate := 0.0
		if cacheTotal > 0 {
			hitRate = float64(agg.cacheRead) / float64(cacheTotal) * 100
		}
		cpm := 0.0
		if agg.messages > 0 {
			cpm = agg.cost / float64(agg.messages)
		}
		avgTokens := 0.0
		if agg.messages > 0 {
			avgTokens = float64(totalTokens) / float64(agg.messages)
		}
		data.ModelComparison = append(data.ModelComparison, model.ModelEfficiency{
			Model:           m,
			Sessions:        agg.sessions,
			TotalCost:       agg.cost,
			TotalMessages:   agg.messages,
			CostPerMessage:  cpm,
			AvgTokensPerMsg: avgTokens,
			CacheHitRate:    hitRate,
			InputTokens:     agg.input,
			OutputTokens:    agg.output,
			CacheReadTokens: agg.cacheRead,
		})
	}
	sort.Slice(data.ModelComparison, func(i, j int) bool {
		return data.ModelComparison[i].TotalCost > data.ModelComparison[j].TotalCost
	})

	// Thinking impact
	thinkAvgCPM := 0.0
	if thinkMsgTotal > 0 {
		thinkAvgCPM = thinkCostTotal / float64(thinkMsgTotal)
	}
	noThinkAvgCPM := 0.0
	if noThinkMsgTotal > 0 {
		noThinkAvgCPM = noThinkCostTotal / float64(noThinkMsgTotal)
	}
	thinkAvgCost := 0.0
	if len(thinkSessions) > 0 {
		thinkAvgCost = thinkCostTotal / float64(len(thinkSessions))
	}
	noThinkAvgCost := 0.0
	if len(noThinkSessions) > 0 {
		noThinkAvgCost = noThinkCostTotal / float64(len(noThinkSessions))
	}
	thinkAvgDur := 0.0
	if len(thinkSessions) > 0 {
		thinkAvgDur = thinkDurTotal / float64(len(thinkSessions))
	}
	noThinkAvgDur := 0.0
	if len(noThinkSessions) > 0 {
		noThinkAvgDur = noThinkDurTotal / float64(len(noThinkSessions))
	}
	multiplier := 0.0
	if noThinkAvgCPM > 0 {
		multiplier = thinkAvgCPM / noThinkAvgCPM
	}
	data.ThinkingImpact = model.ThinkingImpact{
		WithThinking: model.EfficiencyGroup{
			Sessions:      len(thinkSessions),
			AvgCost:       thinkAvgCost,
			AvgCostPerMsg: thinkAvgCPM,
			TotalCost:     thinkCostTotal,
			AvgDuration:   thinkAvgDur,
		},
		WithoutThinking: model.EfficiencyGroup{
			Sessions:      len(noThinkSessions),
			AvgCost:       noThinkAvgCost,
			AvgCostPerMsg: noThinkAvgCPM,
			TotalCost:     noThinkCostTotal,
			AvgDuration:   noThinkAvgDur,
		},
		CostMultiplier: multiplier,
	}

	// Vampire sessions (top 10 by cost)
	sort.Slice(vampires, func(i, j int) bool {
		return vampires[i].session.EstimatedCost > vampires[j].session.EstimatedCost
	})
	limit := 10
	if len(vampires) < limit {
		limit = len(vampires)
	}
	for _, v := range vampires[:limit] {
		s := v.session
		primaryModel := "unknown"
		var maxT int64
		for m, tu := range s.ModelTokens {
			if tu.Total() > maxT {
				maxT = tu.Total()
				primaryModel = m
			}
		}
		totalInput := s.TotalTokens.InputTokens + s.TotalTokens.CacheReadTokens + s.TotalTokens.CacheWriteTokens
		hitRate := 0.0
		if totalInput > 0 {
			hitRate = float64(s.TotalTokens.CacheReadTokens) / float64(totalInput) * 100
		}
		prompt := s.FirstPrompt
		if len(prompt) > 100 {
			prompt = prompt[:100]
		}
		data.VampireSessions = append(data.VampireSessions, model.VampireSession{
			SessionID:     s.SessionID,
			Slug:          s.Slug,
			FirstPrompt:   prompt,
			Cost:          s.EstimatedCost,
			Messages:      s.TotalMsgCount,
			CostPerMsg:    v.cpm,
			Duration:      s.DurationMin,
			PrimaryModel:  primaryModel,
			HasThinking:   s.HasThinking,
			HasCompaction: s.HasCompaction,
			CacheHitRate:  hitRate,
			StartTime:     s.StartTime.Format(time.RFC3339),
		})
	}

	// Cache by session (worst 20 by hit rate, min 5 messages to be meaningful)
	sort.Slice(cacheCandidates, func(i, j int) bool {
		return cacheCandidates[i].hitRate < cacheCandidates[j].hitRate
	})
	cacheLimit := 20
	count := 0
	for _, c := range cacheCandidates {
		if count >= cacheLimit {
			break
		}
		if c.session.TotalMsgCount < 5 {
			continue
		}
		prompt := c.session.FirstPrompt
		if len(prompt) > 80 {
			prompt = prompt[:80]
		}
		data.CacheBySession = append(data.CacheBySession, model.SessionCacheInfo{
			SessionID:    c.session.SessionID,
			Slug:         c.session.Slug,
			FirstPrompt:  prompt,
			CacheHitRate: math.Round(c.hitRate*10) / 10,
			Cost:         c.session.EstimatedCost,
			Messages:     c.session.TotalMsgCount,
			WastedTokens: c.wasted,
		})
		count++
	}

	// Health score (0-100)
	data.HealthScore = computeHealthScore(data)

	return data
}

func percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := float64(p) / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))
	if lower == upper {
		return sorted[lower]
	}
	frac := idx - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

func buildCostDistribution(cpms []float64) []model.CostBucket {
	buckets := []model.CostBucket{
		{Label: "$0-0.01", Min: 0, Max: 0.01},
		{Label: "$0.01-0.05", Min: 0.01, Max: 0.05},
		{Label: "$0.05-0.10", Min: 0.05, Max: 0.10},
		{Label: "$0.10-0.25", Min: 0.10, Max: 0.25},
		{Label: "$0.25-0.50", Min: 0.25, Max: 0.50},
		{Label: "$0.50-1.00", Min: 0.50, Max: 1.00},
		{Label: "$1.00+", Min: 1.00, Max: math.MaxFloat64},
	}
	for _, cpm := range cpms {
		for i := range buckets {
			if cpm >= buckets[i].Min && cpm < buckets[i].Max {
				buckets[i].Count++
				break
			}
		}
	}
	return buckets
}

func computeHealthScore(data *model.EfficiencyData) int {
	score := 100

	// Penalize if cache hit rate is low across models
	for _, m := range data.ModelComparison {
		if m.CacheHitRate < 30 && m.Sessions > 5 {
			score -= 15
		} else if m.CacheHitRate < 50 && m.Sessions > 5 {
			score -= 8
		}
	}

	// Penalize if thinking multiplier is too high
	if data.ThinkingImpact.CostMultiplier > 3.0 {
		score -= 15
	} else if data.ThinkingImpact.CostMultiplier > 2.0 {
		score -= 8
	}

	// Penalize if P99 cost-per-message is extreme
	if data.CostPerMessage.Median > 0 {
		ratio := data.CostPerMessage.P99 / data.CostPerMessage.Median
		if ratio > 20 {
			score -= 15
		} else if ratio > 10 {
			score -= 8
		}
	}

	if score < 0 {
		score = 0
	}
	return score
}

// ExportForCache exports sessions and file states for persistence.
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

// containsCI performs a case-insensitive search.
func containsCI(haystack, needle string) bool {
	h := strings.ToLower(haystack)
	n := strings.ToLower(needle)
	return strings.Contains(h, n)
}
