package service

import (
	"github.com/zhaiyz/ocsession/internal/config"
	"github.com/zhaiyz/ocsession/internal/store"
)

// SessionService handles session business logic
type SessionService struct {
	store    store.Store
	config   *config.Config
	sessions []store.Session
}

// NewSessionService creates a new session service
func NewSessionService(store store.Store, cfg *config.Config) *SessionService {
	return &SessionService{
		store:  store,
		config: cfg,
	}
}

// LoadSessions loads sessions from store
func (s *SessionService) LoadSessions() error {
	sessions, err := s.store.LoadSessions()
	if err != nil {
		return err
	}
	s.sessions = sessions
	return nil
}

// GetAllSessions returns all loaded sessions
func (s *SessionService) GetAllSessions() []store.Session {
	return s.sessions
}

// GetSessionDetail returns session detail
func (s *SessionService) GetSessionDetail(id string) (*store.SessionDetail, error) {
	return s.store.GetSessionDetail(id)
}
