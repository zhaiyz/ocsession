package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

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

// GetSessionDetail loads session detail with messages and stats
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

	// Query message count
	var messageCount int
	countQuery := `SELECT COUNT(*) FROM message WHERE session_id = ?`
	s.db.QueryRow(countQuery, id).Scan(&messageCount)

	// Query user messages from part table (joined with message to filter role)
	partQuery := `
		SELECT p.data
		FROM part p
		JOIN message m ON p.message_id = m.id
		WHERE p.session_id = ?
		  AND json_extract(m.data, '$.role') = 'user'
		  AND json_extract(p.data, '$.type') = 'text'
		ORDER BY p.time_created ASC
	`

	rows, err := s.db.Query(partQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query parts: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var data string
		err := rows.Scan(&data)
		if err != nil {
			continue
		}

		// Parse JSON to extract text
		var partData struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}

		if err := json.Unmarshal([]byte(data), &partData); err != nil {
			continue
		}

		// Only include user text messages
		if partData.Type == "text" && partData.Text != "" {
			messages = append(messages, Message{Content: partData.Text})
		}
	}

	// Keep all messages (don't truncate)
	// Truncation will be done in the UI layer

	// Calculate session duration
	duration := int64(0)
	if sess.Updated > 0 && sess.Created > 0 {
		duration = (sess.Updated - sess.Created) / 1000 // milliseconds to seconds
	}

	return &SessionDetail{
		Session:      sess,
		LastMessages: messages,
		Stats: SessionStats{
			MessageCount: messageCount,
			Duration:     duration,
		},
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
