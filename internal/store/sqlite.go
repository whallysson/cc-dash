package store

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/whallysson/cc-dash/internal/model"

	_ "modernc.org/sqlite"
)

// Store persists the session index in SQLite for warm starts.
type Store struct {
	db *sql.DB
	mu sync.Mutex // serialize all writes to avoid SQLITE_BUSY
}

// Open opens or creates the cache database.
func Open(claudeDir string) (*Store, error) {
	dbPath := filepath.Join(claudeDir, ".cc-dash-cache.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA busy_timeout=5000")

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			slug TEXT NOT NULL,
			data_json TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS file_states (
			path TEXT PRIMARY KEY,
			mtime TEXT NOT NULL,
			size INTEGER NOT NULL,
			offset INTEGER NOT NULL,
			session_id TEXT NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_slug ON sessions(slug);
	`)
	if err != nil {
		return err
	}
	// Migrate existing table: add session_id column if missing
	s.db.Exec("ALTER TABLE file_states ADD COLUMN session_id TEXT NOT NULL DEFAULT ''")
	return nil
}

// LoadSessions loads all sessions from cache.
func (s *Store) LoadSessions() (map[string]*model.SessionMeta, map[string]model.FileState, error) {
	sessions := make(map[string]*model.SessionMeta)
	fileStates := make(map[string]model.FileState)

	rows, err := s.db.Query("SELECT session_id, data_json FROM sessions")
	if err != nil {
		return sessions, fileStates, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, dataJSON string
		if err := rows.Scan(&id, &dataJSON); err != nil {
			continue
		}
		var meta model.SessionMeta
		if err := json.Unmarshal([]byte(dataJSON), &meta); err != nil {
			continue
		}
		sessions[id] = &meta
	}

	rows2, err := s.db.Query("SELECT path, mtime, size, offset, COALESCE(session_id, '') FROM file_states")
	if err != nil {
		return sessions, fileStates, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var path, mtimeStr, sessionID string
		var size, offset int64
		if err := rows2.Scan(&path, &mtimeStr, &size, &offset, &sessionID); err != nil {
			continue
		}
		mtime, _ := time.Parse(time.RFC3339Nano, mtimeStr)
		fileStates[path] = model.FileState{
			Path:      path,
			Mtime:     mtime,
			Size:      size,
			Offset:    offset,
			SessionID: sessionID,
		}
	}

	return sessions, fileStates, nil
}

// SaveSession persists a session to cache.
func (s *Store) SaveSession(meta *model.SessionMeta) {
	data, err := json.Marshal(meta)
	if err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = s.db.Exec(
		`INSERT OR REPLACE INTO sessions (session_id, slug, data_json, updated_at) VALUES (?, ?, ?, ?)`,
		meta.SessionID, meta.Slug, string(data), time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		log.Printf("[store] failed to save session %s: %v", meta.SessionID, err)
	}
}

// SaveFileState persists a file state to cache.
func (s *Store) SaveFileState(fs model.FileState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO file_states (path, mtime, size, offset, session_id) VALUES (?, ?, ?, ?, ?)`,
		fs.Path, fs.Mtime.Format(time.RFC3339Nano), fs.Size, fs.Offset, fs.SessionID,
	)
	if err != nil {
		log.Printf("[store] failed to save file state %s: %v", filepath.Base(fs.Path), err)
	}
}

// SaveBatch saves multiple sessions and file states at once.
func (s *Store) SaveBatch(sessions map[string]*model.SessionMeta, fileStates map[string]model.FileState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmtSession, _ := tx.Prepare(`INSERT OR REPLACE INTO sessions (session_id, slug, data_json, updated_at) VALUES (?, ?, ?, ?)`)
	stmtFile, _ := tx.Prepare(`INSERT OR REPLACE INTO file_states (path, mtime, size, offset, session_id) VALUES (?, ?, ?, ?, ?)`)

	now := time.Now().UTC().Format(time.RFC3339)

	for id, meta := range sessions {
		data, err := json.Marshal(meta)
		if err != nil {
			continue
		}
		stmtSession.Exec(id, meta.Slug, string(data), now)
	}

	for _, fs := range fileStates {
		stmtFile.Exec(fs.Path, fs.Mtime.Format(time.RFC3339Nano), fs.Size, fs.Offset, fs.SessionID)
	}

	stmtSession.Close()
	stmtFile.Close()
	tx.Commit()
}

// Exists checks if the database exists and has data.
func Exists(claudeDir string) bool {
	dbPath := filepath.Join(claudeDir, ".cc-dash-cache.db")
	info, err := os.Stat(dbPath)
	return err == nil && info.Size() > 0
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}
