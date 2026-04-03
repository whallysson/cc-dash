package store

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/whallysson/cc-dash/internal/model"

	_ "modernc.org/sqlite"
)

// Store persiste o índice de sessões em SQLite para warm starts.
type Store struct {
	db *sql.DB
}

// Open abre ou cria o banco de cache.
func Open(claudeDir string) (*Store, error) {
	dbPath := filepath.Join(claudeDir, ".cc-dash-cache.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// WAL mode para leituras concorrentes
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")

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
			offset INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_slug ON sessions(slug);
	`)
	return err
}

// LoadSessions carrega todas as sessões do cache.
func (s *Store) LoadSessions() (map[string]*model.SessionMeta, map[string]model.FileState, error) {
	sessions := make(map[string]*model.SessionMeta)
	fileStates := make(map[string]model.FileState)

	// Carregar sessões
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

	// Carregar file states
	rows2, err := s.db.Query("SELECT path, mtime, size, offset FROM file_states")
	if err != nil {
		return sessions, fileStates, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var path, mtimeStr string
		var size, offset int64
		if err := rows2.Scan(&path, &mtimeStr, &size, &offset); err != nil {
			continue
		}
		mtime, _ := time.Parse(time.RFC3339Nano, mtimeStr)
		fileStates[path] = model.FileState{
			Path:   path,
			Mtime:  mtime,
			Size:   size,
			Offset: offset,
		}
	}

	return sessions, fileStates, nil
}

// SaveSession persiste uma sessão no cache.
func (s *Store) SaveSession(meta *model.SessionMeta) {
	data, err := json.Marshal(meta)
	if err != nil {
		return
	}
	_, err = s.db.Exec(
		`INSERT OR REPLACE INTO sessions (session_id, slug, data_json, updated_at) VALUES (?, ?, ?, ?)`,
		meta.SessionID, meta.Slug, string(data), time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		log.Printf("[store] erro ao salvar sessão %s: %v", meta.SessionID, err)
	}
}

// SaveFileState persiste o estado de um arquivo no cache.
func (s *Store) SaveFileState(fs model.FileState) {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO file_states (path, mtime, size, offset) VALUES (?, ?, ?, ?)`,
		fs.Path, fs.Mtime.Format(time.RFC3339Nano), fs.Size, fs.Offset,
	)
	if err != nil {
		log.Printf("[store] erro ao salvar file state %s: %v", filepath.Base(fs.Path), err)
	}
}

// SaveBatch salva múltiplas sessões e file states de uma vez.
func (s *Store) SaveBatch(sessions map[string]*model.SessionMeta, fileStates map[string]model.FileState) {
	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmtSession, _ := tx.Prepare(`INSERT OR REPLACE INTO sessions (session_id, slug, data_json, updated_at) VALUES (?, ?, ?, ?)`)
	stmtFile, _ := tx.Prepare(`INSERT OR REPLACE INTO file_states (path, mtime, size, offset) VALUES (?, ?, ?, ?)`)

	now := time.Now().UTC().Format(time.RFC3339)

	for id, meta := range sessions {
		data, err := json.Marshal(meta)
		if err != nil {
			continue
		}
		stmtSession.Exec(id, meta.Slug, string(data), now)
	}

	for _, fs := range fileStates {
		stmtFile.Exec(fs.Path, fs.Mtime.Format(time.RFC3339Nano), fs.Size, fs.Offset)
	}

	stmtSession.Close()
	stmtFile.Close()
	tx.Commit()
}

// Exists verifica se o banco existe e tem dados.
func Exists(claudeDir string) bool {
	dbPath := filepath.Join(claudeDir, ".cc-dash-cache.db")
	info, err := os.Stat(dbPath)
	return err == nil && info.Size() > 0
}

// Close fecha o banco.
func (s *Store) Close() error {
	return s.db.Close()
}
