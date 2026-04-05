package store

import (
    "database/sql"
    "fmt"
    "os"
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
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".local", "share", "opencode", "opencode.db")
}

// LoadSessions loads all sessions from database
func (s *SQLiteStore) LoadSessions() ([]Session, error) {
    query := `
        SELECT id, title, time_updated, time_created, project_id, directory
        FROM session
        WHERE parent_id IS NULL
        ORDER BY time_updated DESC
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
        SELECT id, title, time_updated, time_created, project_id, directory
        FROM session
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
    
    // Query last messages (simplified - just get message data)
    msgQuery := `
        SELECT data
        FROM message
        WHERE session_id = ?
        ORDER BY time_created DESC
        LIMIT 10
    `
    
    rows, err := s.db.Query(msgQuery, id)
    if err != nil {
        return nil, fmt.Errorf("failed to query messages: %w", err)
    }
    defer rows.Close()
    
    var messages []Message
    for rows.Next() {
        var data string
        err := rows.Scan(&data)
        if err != nil {
            return nil, fmt.Errorf("failed to scan message: %w", err)
        }
        // Create a message with the data content
        messages = append(messages, Message{Content: data})
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
        SELECT id, title, time_updated, time_created, project_id, directory
        FROM session
        WHERE parent_id IS NULL
          AND (title LIKE ? OR directory LIKE ?)
        ORDER BY time_updated DESC
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
