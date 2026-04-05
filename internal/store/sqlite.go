package store

import (
    "database/sql"
    "fmt"
    "path/filepath"
    
    _ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements Store interface using SQLite database
type SQLiteStore struct {
    db *sql.DB
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    return &SQLiteStore{db: db}, nil
}

// DefaultDBPath returns the default OpenCode database path
func DefaultDBPath() string {
    homeDir, _ := filepath.HomeDir()
    return filepath.Join(homeDir, ".local", "share", "opencode", "opencode.db")
}

// LoadSessions loads all sessions from database
func (s *SQLiteStore) LoadSessions() ([]Session, error) {
    query := `
        SELECT id, title, updated, created, project_id, directory
        FROM sessions
        ORDER BY updated DESC
        LIMIT 50
    `
    
    rows, err := s.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to query sessions: %w", err)
    }
    defer rows.Close()
    
    var sessions []Session
    for rows.Next() {
        var sess Session
        err := rows.Scan(
            &sess.ID,
            &sess.Title,
            &sess.Updated,
            &sess.Created,
            &sess.ProjectID,
            &sess.Directory,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan session: %w", err)
        }
        sessions = append(sessions, sess)
    }
    
    return sessions, nil
}

// GetSessionDetail loads session detail with messages
func (s *SQLiteStore) GetSessionDetail(id string) (*SessionDetail, error) {
    // Query session basic info
    sessQuery := `
        SELECT id, title, updated, created, project_id, directory
        FROM sessions
        WHERE id = ?
    `
    
    var sess Session
    err := s.db.QueryRow(sessQuery, id).Scan(
        &sess.ID,
        &sess.Title,
        &sess.Updated,
        &sess.Created,
        &sess.ProjectID,
        &sess.Directory,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query session: %w", err)
    }
    
    // Query last messages
    msgQuery := `
        SELECT content, timestamp, role
        FROM messages
        WHERE session_id = ?
        ORDER BY timestamp DESC
        LIMIT 10
    `
    
    rows, err := s.db.Query(msgQuery, id)
    if err != nil {
        return nil, fmt.Errorf("failed to query messages: %w", err)
    }
    defer rows.Close()
    
    var messages []Message
    for rows.Next() {
        var msg Message
        err := rows.Scan(&msg.Content, &msg.Timestamp, &msg.Role)
        if err != nil {
            return nil, fmt.Errorf("failed to scan message: %w", err)
        }
        messages = append(messages, msg)
    }
    
    return &SessionDetail{
        Session:      sess,
        LastMessages: messages,
        Stats:        SessionStats{}, // TODO: implement stats calculation
    }, nil
}

// SearchSessions searches sessions by query
func (s *SQLiteStore) SearchSessions(query string) ([]Session, error) {
    searchQuery := `
        SELECT id, title, updated, created, project_id, directory
        FROM sessions
        WHERE title LIKE ? OR directory LIKE ?
        ORDER BY updated DESC
        LIMIT 50
    `
    
    searchTerm := "%" + query + "%"
    rows, err := s.db.Query(searchQuery, searchTerm, searchTerm)
    if err != nil {
        return nil, fmt.Errorf("failed to search sessions: %w", err)
    }
    defer rows.Close()
    
    var sessions []Session
    for rows.Next() {
        var sess Session
        err := rows.Scan(
            &sess.ID,
            &sess.Title,
            &sess.Updated,
            &sess.Created,
            &sess.ProjectID,
            &sess.Directory,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan session: %w", err)
        }
        sessions = append(sessions, sess)
    }
    
    return sessions, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
    return s.db.Close()
}
