package store

import "time"

type Session struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Created   time.Time `json:"created"`
    Updated   time.Time `json:"updated"`
    ProjectID string    `json:"project_id"`
    Directory string    `json:"directory"`
    
    // Extended fields from config
    Tags      []string
    Alias     string
    Notes     string
}

type SessionDetail struct {
    Session      Session
    LastMessages []Message
    Stats        SessionStats
}

type SessionStats struct {
    TokenCount    int
    MessageCount  int
    ToolCallCount int
    Cost          float64
}

type FilterCriteria struct {
    Tags       []string
    Project    string
    DateFrom   time.Time
    DateTo     time.Time
    Query      string
}
