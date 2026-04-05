package store

// Store defines the interface for session data access
type Store interface {
    LoadSessions() ([]Session, error)
    GetSessionDetail(id string) (*SessionDetail, error)
    SearchSessions(query string) ([]Session, error)
    Close() error
}
